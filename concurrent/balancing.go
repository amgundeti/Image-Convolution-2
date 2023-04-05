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

// NewWorkBalancingExecutor returns an ExecutorService that is implemented using the work-balancing algorithm.
// @param capacity - The number of goroutines in the pool
// @param threshold - The number of items that a goroutine in the pool can
// grab from the executor in one time period. For example, if threshold = 10
// this means that a goroutine can grab 10 items from the executor all at
// once to place into their local queue before grabbing more items. It's
// not required that you use this parameter in your implementation.
// @param thresholdBalance - The threshold used to know when to perform
// balancing. Remember, if two local queues are to be balanced the
// difference in the sizes of the queues must be greater than or equal to
// thresholdBalance. You must use this parameter in your implementation.

type WorkBalancingExecutor struct {
	Capacity  int
	threshold int
	thresholdBalance int

	wg         *sync.WaitGroup
	mu         sync.Mutex
	done       bool
	totalTasks int
	queues     []DEQueue
	queuesMus	[]sync.Mutex
	currIndex  int

	tasksPushed int
}

func NewWorkBalancingExecutor(capacity int, threshold int, thresholdBalance int) ExecutorService {
	
	executor := &WorkBalancingExecutor{Capacity: capacity, threshold: threshold, thresholdBalance: thresholdBalance,
		wg: &sync.WaitGroup{}, done: false, totalTasks: 0, currIndex: 0, tasksPushed: 0}

	//Make Queues and Slice of Mutexes
	queues := make([]DEQueue, capacity)
	for i := 0; i < capacity; i++ {
		queues[i] = NewUnBoundedDEQueue()
	}

	queuesMus := make([]sync.Mutex, capacity)
	for i:=0; i < capacity; i++{
		queuesMus[i] = sync.Mutex{}
	}

	executor.queues = queues
	executor.queuesMus = queuesMus
	return executor

}

//////////////////////////////////////////////////////////////
// W O  R K | B A L A N C I N G | E X E C U T O R | F U N C S
//////////////////////////////////////////////////////////////

func (w *WorkBalancingExecutor) Shutdown() {
	w.mu.Lock()
	w.done = true
	w.mu.Unlock()
}

func (w *WorkBalancingExecutor) Submit(task interface{}) Future {

	w.mu.Lock()
	w.queues[w.currIndex].PushBottom(task)
	w.currIndex = (w.currIndex + 1) % w.Capacity
	w.totalTasks += 1
	w.tasksPushed += 1
	w.mu.Unlock()
	return nil
}

//////////////////////////////////////////////////////////////
// W O  R K | B A L A N C I N G | M A I N | F U N C 
//////////////////////////////////////////////////////////////

func WorkBalancing(config Config) {

	effectsPathFile := fmt.Sprintf("../proj3/data/effects.txt")
	effectsFile, err := os.Open(effectsPathFile)

	if err != nil {
		print(err)
	}

	reader := json.NewDecoder(effectsFile)
	directories := strings.Split(config.DataDirs, "+")

	//create executor and launch routines
	executor := (NewWorkBalancingExecutor(config.ThreadCount, 2, config.ThresholdBalance)).(*WorkBalancingExecutor)
	executor.wg.Add(executor.Capacity)
	for i := 0; i < executor.Capacity; i++ {
		go BalanceWorker(executor, i)
	}

	//decode json
	for reader.More() {
		req := Request{}
		err := reader.Decode(&req)

		if err != nil {
			print(err)
			return
		}

		//pass tasks as *png.ImageTask
		for _, directory := range directories {
			filePath := "../proj3/data/in/" + directory + "/" + req.InPath
			pngImg, _ := png.Load(filePath)
			pngImg.Effects = req.Effects
			outname := directory + "_" + req.OutPath
			pngImg.OutName = outname
			executor.Submit(pngImg)
		}
	}

	//set shutdown and wait
	executor.Shutdown()
	executor.wg.Wait()
}


//////////////////////////////////////////////////////////////
// W O  R K | B A L A N C I N G | W O R K E R | F U N C
//////////////////////////////////////////////////////////////

func BalanceWorker(w *WorkBalancingExecutor, id int){

	for{

		myQLen := w.queues[id].QLen()
		task := w.queues[id].PopBottom()

		//check if queue is returning tasks
		if checkedTask, ok := task.(*png.ImageTask); ok {
			checkedTask.Run()
			w.mu.Lock()
			w.totalTasks -= 1
			w.mu.Unlock()
		}

		s1 := rand.NewSource(time.Now().UnixNano())
		r1 := rand.New(s1)
		
		//check queue length against random generator
		myQLen = w.queues[id].QLen()
		randomBalance := r1.Intn(myQLen + 1)

		if (randomBalance == myQLen){

			s2 := rand.NewSource(time.Now().UnixNano())
			r2 := rand.New(s2)

			//find target and attempt to balance
			balancingTarget := r2.Intn(w.Capacity)

			if balancingTarget == id{
				continue
			}

			first, second := setThreadOrder(id, balancingTarget)

			//acquire "outside" locks in index order
			w.queuesMus[first].Lock()
			w.queuesMus[second].Lock()
			balanceTasks(w, first, second)
			w.queuesMus[second].Unlock()
			w.queuesMus[first].Unlock()
		}

		//check if time to return
		w.mu.Lock()
		if w.done && (w.totalTasks == 0) && (w.queues[id].IsEmpty()) {
			w.wg.Done()
			w.mu.Unlock()
			return
		} 
		w.mu.Unlock()
	}
}

//////////////////////////////////////////////////////////////////////////
// W O  R K | B A L A N C I N G | W O R K E R | S U P P O R T | F U N C S
//////////////////////////////////////////////////////////////////////////

//Sets thread order for "outside" mutex acquisition
func setThreadOrder(first int, second int) (int, int){

	if first > second{
		return first, second
	} else{
		return second, first
	}

}

//attempts rebalance

func balanceTasks(w *WorkBalancingExecutor, first int, second int ){

	var qMax int
	var qMin int
	
	//determine which queue is longer
	if w.queues[first].QLen() < w.queues[second].QLen(){
		qMin = first
		qMax = second
	}  else{
		qMin = second
		qMax = first
	}

	diff := w.queues[qMax].QLen() - w.queues[qMin].QLen() 

	if diff > w.thresholdBalance{
		for w.queues[qMin].QLen() < w.queues[qMax].QLen(){

			task := w.queues[qMax].PopTop()
			w.queues[qMin].PushBottom(task)

		}
	}
}

