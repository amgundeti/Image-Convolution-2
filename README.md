# Image Convolution - 2

# Overview

This project is a reimplementation of the the image convolution program with work stealing and work balancing methods. In short, the program takes command line arguments, and initializes the associated method program (e.g., work stealing vs. work balancing). The method program then reads in a json file, loads images from disk, and applies a combination of greyscale, edge detection, sharpening, and blurring effects according to the method specified. 


---

# Parallel Implementations

Both work stealing and work balancing methods are supported by an unbounded dequeue. The queue is a coarse grained double-ended queue that uses sentinel nodes. The dequeue supports push bottom, pop bottom, and pop top methods and has support functions like QLen (for extracting length of queue) and IsEmpty. 

Tasks constitute of a whole image, applying all associated effects, and saving the image. As such, all tasks are submitted as runables because no return types are expected.

Both parallel implementations are coordinated by an overarching function (WorkStealing and WorkBalancing). These functions set up the json decoder, create an instance of the executor, and spawn the required go routines. The executor distributes tasks across worker queues one at a time and workers pop from their queues. The implementation protects against “empty dequeues” by relying on Go’s type casting. If the popped item cannot be type cast to *png.ImageTask, the routine continues to the next iteration.

Synchronization across the queues is generally maintained via mutexes. As mentioned above, each queue is a coarse grained dequeue and the executor is a mutex protected struct. The worker routines continue to run as long as the totalTasks counter is above 0, and the done flag is false.

*Below are specifics for work stealing and work balancing.*

### Work Stealing

In Work Stealing, workers first try to dequeue from their queues. If successful in typecasting to *png.ImageTask, the worker calls Run() on the object and moves onto the next iteration of the for-loop. If a worker’s queue is empty, the worker randomly targets another thread and attempts to steal work by calling PopTop() on their queue. Randomization is supported by Go’s native random number generator. 

### Work Balancing

In Work Balancing, workers also dequeue from their queues first. Whether successful or not, they then generate a random number between 0 and the length of their queue (inclusive). If this random number matches the length of their queue, they attempt to rebalance a random thread’s queue. 

To support this, the executor maintains a slice of mutexes (the queuesMus slice) that correspond to each worker queue. These mutexes stop two threads from trying to rebalance the same queue or each other. The implementation follows the recommendations of “The Art of Multicore Programming” and locks queues in index order to avoid deadlocking. Furthermore, the balanceTasks function takes on the work of popping tasks from one queue and pushing into another queue. 

---

# Challenges

Challenges were limited to understanding the work stealing and work balancing methods. In addition, setting up the unbounded dequeue was particularly challenging because it is possible to push and pop nil values from the queue in both implementations. Ultimately including a sentinel node on queue initialization cleared roadblocks presented by returning a nil head and tail. 

For the implementations themselves, it was difficult to settle on the right unit of work. Initially I hoped to parcel work at the pixel level, however mutex contention would have made this impractical. I also attempted to parcel work as sections of images, however, this implementation was also impractical because application of effects must be coordinated. While feasible, it seemed to diminish any benefits of work stealing or work balancing.

---

# Performance Analysis
![Speedup Work Stealing](https://user-images.githubusercontent.com/107568169/230180046-348f97fa-c218-4059-bbc6-152e324e03c1.png)
![speedup-balancing](https://user-images.githubusercontent.com/107568169/230180203-10fa2a84-8962-4e63-b9f3-87de80c8b5d1.png)


The above speed up graphs are generated on an M1 MacBook Pro with 10 cores (the graphs go up to 12 cores to support comparison with project 2 implementation). The work stealing and work balancing implementations show peak speed up of ~3.5x compared to ~6.5x for pipeline and BSP. This is somewhat expected given the bottlenecks mutexes impose. BSP and and pipeline are mutex free implementations, and arguably are better suited for the task at hand. Pushing a whole image from one queue to another does little to rebalance loads. 

The speed up in these graphs is also limited by the size of the total task. The small, big, and mixture folders only hold 10 images each. It is possible speed up would have been higher if the queue sizes could grow into the 10s or 100s. While certainly possible to expand the data set, time constrains were such that I could not.  Work stealing appears to be the primary victim of a small data set because after threads finish their tasks they continue to “spam” the mutexes of other threads increasing contention. This explains the sharp drop off in speed up after 4-6 cores. Certainly in the case of 12 cores, 2 threads never receive tasks and create tremendous contention in the program. For example, threads recorded several hundred thousand attempts at acquiring a mutex in testing runs for the big folder. Work balancing experiences slightly less lock contention thanks to randomization.

The data set size also explains the plateau in speed up for Work Balancing. The tests were run with a ThresholdBalance of 2, and with more than 4 threads there are almost no instances where a thread would find an opportunity to work balance. The variation in speed up for the mixture folder is likely due to cpu traffic on the machine - 2 runs were about 10% slower than the other 3.

One phenomenon that continues to puzzle me is the low speed up for the mixture folder. In project 2 I hypothesized work stealing/work balancing would improve speed up here as threads finish work at different rates based on the image size. However, it has not played out. It is possible the task is not defined at the right level for work stealing and balancing to have any effect on this folder. Project 2 speed up graphs are included below for reference.

