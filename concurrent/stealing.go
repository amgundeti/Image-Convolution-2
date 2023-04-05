package concurrent

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"proj3/png"
	"strings"
	"sync"
	"time"
)

// NewWorkStealingExecutor returns an ExecutorService that is implemented using the work-stealing algorithm.
// @param Capacity - The number of goroutines in the pool
// @param threshold - The number of items that a goroutine in the pool can
// grab from the executor in one time period. For example, if threshold = 10
// this means that a goroutine can grab 10 items from the executor all at
// once to place into their local queue before grabbing more items. It's
// not required that you use this parameter in your implementation.

type WorkStealExecutorService struct {
	Capacity  int
	threshold int

	wg         *sync.WaitGroup
	mu         sync.Mutex
	done       bool
	totalTasks int
	queues     []DEQueue
	currIndex  int

	tasksPushed int
}

func NewWorkStealingExecutor(Capacity, threshold int) ExecutorService {

	executor := &WorkStealExecutorService{Capacity: Capacity, threshold: threshold, wg: &sync.WaitGroup{},
		done: false, totalTasks: 0, currIndex: 0, tasksPushed: 0}

	queues := make([]DEQueue, Capacity)
	for i := 0; i < Capacity; i++ {
		queues[i] = NewUnBoundedDEQueue()
	}
	executor.queues = queues
	return executor
}

//////////////////////////////////////////////////////////////
// W O  R K | B A L A N C I N G | E X E C U T O R | F U N C S
//////////////////////////////////////////////////////////////

func (w *WorkStealExecutorService) Shutdown() {
	w.mu.Lock()
	w.done = true
	w.mu.Unlock()
}

func (w *WorkStealExecutorService) Submit(task interface{}) Future {

	w.mu.Lock()
	w.queues[w.currIndex].PushBottom(task)
	w.currIndex = (w.currIndex + 1) % w.Capacity
	w.totalTasks += 1
	w.tasksPushed += 1
	w.mu.Unlock()
	return nil
}

//////////////////////////////////////////////////////////////
// W O  R K | S T E A L I N G | M A I N | F U N C 
//////////////////////////////////////////////////////////////

func WorkStealing(config Config) {

	effectsPathFile := fmt.Sprintf("../proj3/data/effects.txt")
	effectsFile, err := os.Open(effectsPathFile)

	if err != nil {
		print(err)
	}

	reader := json.NewDecoder(effectsFile)
	directories := strings.Split(config.DataDirs, "+")

	executor := (NewWorkStealingExecutor(config.ThreadCount, 2)).(*WorkStealExecutorService)
	executor.wg.Add(executor.Capacity)
	for i := 0; i < executor.Capacity; i++ {

		go stealWorker(executor, i)
	}

	for reader.More() {
		req := Request{}
		err := reader.Decode(&req)

		if err != nil {
			print(err)
			return
		}

		for _, directory := range directories {
			filePath := "../proj3/data/in/" + directory + "/" + req.InPath
			pngImg, _ := png.Load(filePath)
			pngImg.Effects = req.Effects
			outname := directory + "_" + req.OutPath
			pngImg.OutName = outname
			executor.Submit(pngImg)
		}
	}

	executor.Shutdown()
	executor.wg.Wait()
}

//////////////////////////////////////////////////////////////
// W O  R K | S T E A L I N G | W O R K E R | F U N C
//////////////////////////////////////////////////////////////

func stealWorker(w *WorkStealExecutorService, id int) {

	for {
		selfQLen := w.queues[id].QLen()

		//Do my own work first
		if w.queues[id].QLen() != 0 {

			task := w.queues[id].PopBottom()
			if checkedTask, ok := task.(*png.ImageTask); ok {
				checkedTask.Run()
				w.mu.Lock()
				w.totalTasks -= 1
				w.mu.Unlock()
			}

		} else {
			//steal from other queues if my queue is empty
			//generate random thread to target

			//attribution: GobyExample
			s1 := rand.NewSource(time.Now().UnixNano())
			r1 := rand.New(s1)

			stealIndex := r1.Intn(w.Capacity)

			for stealIndex == id {
				stealIndex = r1.Intn(w.Capacity)
			}

			if !w.queues[stealIndex].IsEmpty(){
				task := w.queues[stealIndex].PopTop()

				if checkedTask, ok := task.(*png.ImageTask); ok {
					checkedTask.Run()
					w.mu.Lock()
					w.totalTasks -= 1
					w.mu.Unlock()
				}
			}
		}
		//check if it is time to return
		w.mu.Lock()
		if w.done && (w.totalTasks == 0) && (w.queues[id].IsEmpty()) {
			w.wg.Done()
			w.mu.Unlock()
			return
		} 
		w.mu.Unlock()
	}
}