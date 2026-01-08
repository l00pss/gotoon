package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/l00pss/gotoon"
)

type Context struct {
	Task     string `toon:"task" json:"task"`
	Location string `toon:"location" json:"location"`
	Season   string `toon:"season" json:"season"`
}

type Hike struct {
	ID            int     `toon:"id" json:"id"`
	Name          string  `toon:"name" json:"name"`
	DistanceKm    float64 `toon:"distanceKm" json:"distanceKm"`
	ElevationGain int     `toon:"elevationGain" json:"elevationGain"`
	Companion     string  `toon:"companion" json:"companion"`
	WasSunny      bool    `toon:"wasSunny" json:"wasSunny"`
}

type HikesData struct {
	Context Context  `toon:"context" json:"context"`
	Friends []string `toon:"friends" json:"friends"`
	Hikes   []Hike   `toon:"hikes" json:"hikes"`
}

func main() {
	fmt.Println("TOON vs JSON Comparison Demo")
	fmt.Println("================================")

	data := HikesData{
		Context: Context{
			Task:     "Our favorite hikes together",
			Location: "Boulder",
			Season:   "spring_2025",
		},
		Friends: []string{"ana", "luis", "sam"},
		Hikes: []Hike{
			{ID: 1, Name: "Blue Lake Trail", DistanceKm: 7.5, ElevationGain: 320, Companion: "ana", WasSunny: true},
			{ID: 2, Name: "Ridge Overlook", DistanceKm: 9.2, ElevationGain: 540, Companion: "luis", WasSunny: false},
			{ID: 3, Name: "Wildflower Loop", DistanceKm: 5.1, ElevationGain: 180, Companion: "sam", WasSunny: true},
		},
	}

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		log.Fatal(err)
	}

	toonData, err := toon.Marshal(data)
	if err != nil {
		log.Fatal(err)
	}

	toonTabData, err := toon.MarshalWithOptions(data, toon.MarshalOptions{
		Indent:     2,
		Delimiter:  toon.DelimiterTab,
		UseTabular: true,
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("\nJSON Format (%d characters):\n", len(jsonData))
	fmt.Printf("%s\n", string(jsonData))

	fmt.Printf("\nTOON Format (%d characters, %.1f%% savings):\n",
		len(toonData), float64(len(jsonData)-len(toonData))/float64(len(jsonData))*100)
	fmt.Printf("%s\n", string(toonData))

	fmt.Printf("\nTOON with Tabs (%d characters, %.1f%% savings):\n",
		len(toonTabData), float64(len(jsonData)-len(toonTabData))/float64(len(jsonData))*100)
	fmt.Printf("%s\n", string(toonTabData))

	fmt.Println("\nRound-trip Test:")
	fmt.Println("===================")

	var decoded HikesData
	if err := toon.Unmarshal(toonData, &decoded); err != nil {
		log.Fatal("Unmarshal failed:", err)
	}

	fmt.Printf("Original context task: %s\n", data.Context.Task)
	fmt.Printf("Decoded context task:  %s\n", decoded.Context.Task)
	fmt.Printf("Original friends count: %d\n", len(data.Friends))
	fmt.Printf("Decoded friends count:  %d\n", len(decoded.Friends))
	fmt.Printf("Original hikes count:   %d\n", len(data.Hikes))
	fmt.Printf("Decoded hikes count:    %d\n", len(decoded.Hikes))

	if decoded.Context.Task == data.Context.Task &&
		len(decoded.Friends) == len(data.Friends) &&
		len(decoded.Hikes) == len(data.Hikes) {
		fmt.Println("Round-trip conversion successful")
	}

	fmt.Println("\nToken Efficiency Analysis:")
	fmt.Println("=================================")

	jsonTokens := len(jsonData) / 4
	toonTokens := len(toonData) / 4
	toonTabTokens := len(toonTabData) / 4

	fmt.Printf("JSON tokens:      ~%d\n", jsonTokens)
	fmt.Printf("TOON tokens:      ~%d (%.1f%% savings)\n",
		toonTokens, float64(jsonTokens-toonTokens)/float64(jsonTokens)*100)
	fmt.Printf("TOON+tabs tokens: ~%d (%.1f%% savings)\n",
		toonTabTokens, float64(jsonTokens-toonTabTokens)/float64(jsonTokens)*100)

	fmt.Printf("\nCost savings at $0.01/1K tokens:\n")
	fmt.Printf("TOON saves:     $%.4f per request\n", float64(jsonTokens-toonTokens)*0.01/1000)
	fmt.Printf("TOON+tabs save: $%.4f per request\n", float64(jsonTokens-toonTabTokens)*0.01/1000)

	fmt.Println("\nFormat Validation:")
	fmt.Println("=====================")

	fmt.Printf("Valid TOON data: %t\n", toon.Valid(toonData))
	fmt.Printf("Valid JSON as TOON: %t\n", toon.Valid(jsonData))
	fmt.Printf("Valid random text: %t\n", toon.Valid([]byte("random invalid text")))
}
