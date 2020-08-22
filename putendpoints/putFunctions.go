package putendpoints

import (
	key "CarbonizeGoRestAPI/apikeys"
	"context"
	"fmt"
	"log"
	"strings"
	"sync"

	"cloud.google.com/go/firestore"
	vision "cloud.google.com/go/vision/apiv1"
	firebase "firebase.google.com/go"
	"google.golang.org/api/option"
)

func getDaysSlice() []string {
	return []string{"Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday", "Sunday"}
}

func getDataChoices() []string {
	return []string{"Carbon", "CalorieIn", "CalorieBurn"}
}

func findMaxKey(filteredResult map[string]float32) string {
	var maxKey string = ""
	var maxValue float32 = 0
	for key, value := range filteredResult {
		if value > maxValue {
			maxKey = key
			maxValue = value
		}
	}
	return maxKey
}

//filters results that are pushed out by Google's Vision API
func filterGoogleResults(foodResults map[string]float32) map[string]float32 {
	empty := struct{}{}
	stringsToElim := map[string]struct{}{"meal": empty, "dish": empty, "food": empty, "good": empty, "cuisine": empty, "noodle": empty, "ingredient": empty,
		"dessert": empty, "icing": empty, "fruit": empty, "plant": empty, "vegetarian food": empty, "art": empty,
		"visual arts": empty, "modern art": empty, "paint": empty, "still life": empty, "illustration": empty, "painting": empty,
		"rectangle": empty, "watercolor paint": empty, "al dente": empty, "knife": empty, "spoon": empty,
		"fork": empty, "plate": empty, "lunch": empty, "breakfast": empty, "dinner": empty, "recipe": empty, "cup": empty, "yellow": empty, "glass": empty}

	//iterate through the food results map
	for key := range foodResults {
		//if key is equal to an item in the stringsToElim map, then delete it from foodResults
		_, ok := stringsToElim[key]
		if ok {
			//key in foodResults needs to be eliminated
			delete(foodResults, key)
		}
	}

	subStringsToElim := [...]string{"dish", "food", "good", "produce", "meal"}
	//eliminate results that have a substring from the subStringsToElim array
	for key := range foodResults {
		for _, str := range subStringsToElim {
			if strings.Contains(key, str) {
				//substring exists in the key, so we should delete the key-value pair and break out of the for loop
				delete(foodResults, key)
				break
			}
		}
	}

	for key, value := range foodResults {
		if key == "pizza cheese" {
			//option is "pizza cheese" which we want to replace with "cheese pizza"
			score := value
			delete(foodResults, key)
			foodResults["cheese pizza"] = score
		}
	}

	scoreReducingWords := [...]string{"noodle"}
	for key, value := range foodResults {
		for _, str := range scoreReducingWords {
			if strings.Contains(key, str) {
				//score reducing word is a substring in the key, so we need to reduce the score of the food
				foodResults[key] = value * 0.875 //multiply value by 7/8 to reduce the score
			}
		}
	}
	return foodResults
}

func convertDataSlice(dataSlice []string) []string {
	convertedSlice := make([]string, 0)
	filterMap := map[string]string{"Carbon Footprint": "Carbon", "Calorie Intake": "CalorieIn", "Calorie Burn": "CalorieBurn"}
	for _, data := range dataSlice {
		convertedSlice = append(convertedSlice, filterMap[data])
	}
	return convertedSlice
}

//AnalyzeImage analyzes the image to find the carbon footprint and
//calorie count of the food in the picture
func AnalyzeImage(id string, downloadURL string) map[string]interface{} {
	ctx := context.Background()
	fmt.Println(downloadURL)
	opt := option.WithCredentialsFile(key.AnalyzeKeyPath)
	//create vision client
	visionClient, err2 := vision.NewImageAnnotatorClient(ctx, opt)
	if err2 != nil {
		log.Fatalln(err2)
	}
	//set the image from the download url that we're given
	image := vision.NewImageFromURI(downloadURL)
	labels, err3 := visionClient.DetectLabels(ctx, image, nil, 10)
	if err3 != nil {
		log.Fatalln(err3)
	}

	googleVisionResults := make(map[string]float32, 0)
	for _, label := range labels {
		fmt.Println(label)
		googleVisionResults[strings.ToLower(label.Description)] = label.Score
	}

	fmt.Println()
	googleVisionResults = filterGoogleResults(googleVisionResults)
	fmt.Println(googleVisionResults)
	bestResult := findMaxKey(googleVisionResults)
	var nextBestResults []string
	for key := range googleVisionResults {
		if key != bestResult {
			nextBestResults = append(nextBestResults, key)
		}
	}
	fmt.Println("Best result:", bestResult)
	fmt.Print("Next best results: ")
	fmt.Println(nextBestResults)
	returningJSONObj := make(map[string]interface{}, 0)
	returningJSONObj["bestResult"] = bestResult
	returningJSONObj["nextBestResults"] = nextBestResults
	fmt.Println(returningJSONObj["bestResult"])
	fmt.Println(returningJSONObj["nextBestResults"])
	return returningJSONObj
}

//ResetGraphData sets all the user's graph data to 0
func ResetGraphData(id string, daysSelection []string, dataSelection []string) bool {
	opt := option.WithCredentialsFile(key.FirebaseKeyPath)
	ctx := context.Background()
	app, err := firebase.NewApp(ctx, nil, opt)
	if err != nil {
		return false
	}
	//connect to client
	client, err := app.Firestore(ctx)
	if err != nil {
		return false
	}
	filteredDataSelection := convertDataSlice(dataSelection)

	//asynchronously reset each path
	var waitGroup sync.WaitGroup
	for _, day := range daysSelection {
		for _, data := range filteredDataSelection {
			//spawn a goroutine to reset the value at that specific path
			waitGroup.Add(1)
			go func(day string, data string) {
				_, err2 := client.Collection("users").Doc(id).Update(ctx, []firestore.Update{
					{Path: day + data, Value: 0},
				})
				if err2 != nil {
					log.Fatal(err2)
				}
				waitGroup.Done()
			}(day, data)
		}
	}
	waitGroup.Wait()
	return true
}
