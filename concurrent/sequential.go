package concurrent

import(
	"fmt"
	"proj3/png"
	"strings"
	"os"
	"encoding/json"
	"image"
)

type Request struct {
	InPath     string
	OutPath    string
	Effects []string
}

func RunSequential(config Config) {

	effectsPathFile := fmt.Sprintf("../proj3/data/effects.txt")
	effectsFile, err := os.Open(effectsPathFile)

	if err != nil{
		print("issue")
	}

	reader := json.NewDecoder(effectsFile)


	directories := strings.Split(config.DataDirs, "+")
	// Work through each effect for each directory for each request

	for reader.More(){
		req := Request{}
		err := reader.Decode(&req)

		if err != nil {
			print(err)
			return
		}

		for _, directory := range directories{

			filePath := "../proj3/data/in/" + directory + "/"+ req.InPath
			pngImg, err := png.Load(filePath)

			if err != nil {
				print(err)
				return
			}

			for _, effect := range req.Effects{
				pngImg.Out = image.NewRGBA64(pngImg.Bounds)
				pngImg.Convolute(effect, pngImg.Bounds.Min.Y, pngImg.Bounds.Max.Y)
				pngImg.In = pngImg.Out
			}
			
			filePath = directory + "_" + req.OutPath
			pngImg.Save(filePath)
		}
	}
}
