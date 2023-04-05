package concurrent

type Config struct {
	DataDirs string //Represents the data directories to use to load the images.
	Mode     string // Represents which scheduler scheme to use
	// If Mode == "s" run the sequential version
	// If Mode == "w" run the workstealing version
	// If Mode == "b" run the work balancing version
	ThreadCount int // Runs the parallel version of the program with the
	// specified number of threads (i.e., goroutines)
	ThresholdBalance int //threshold balancing parameter for work balacning
}

func Schedule(config Config) {
	if config.Mode == "s" {
		RunSequential(config)
	} else if config.Mode == "w" {
		WorkStealing(config)
	}  else if config.Mode == "b"{
		WorkBalancing(config)
	}

}