package main

import (
	"log"
	"net/http"
	"bytes"
	"strings"
	"encoding/json"
	"os"
	"bufio"
	"fmt"
	"regexp"
)

type Country struct {
	Name string `json:"name"`
	Code string `json:"code"`
	GPD  int64  `json:"GPD"`
}

func main() {
	request, err := http.NewRequest("GET", "https://countrycode.org/", nil)
	if err != nil {
		log.Println(err)
	}

	client := &http.Client{}
	resp, err := client.Do(request)

	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)
	s := buf.String()
	openTagIndex := strings.Index(s, "<tbody>")
	closingTagIndex := strings.Index(s, "</tbody>")
	runes := []rune(s)
	resultString := string(runes[openTagIndex:closingTagIndex])

	parseTable(resultString)
	//if err != nil {
	//	panic(err)
	//}
}

func parseTable(input string) {
	finishIndex := strings.Index(input, "</tr>")
	if finishIndex == -1 {
		return
	}
	finishIndex += 4

	//var myCountry Country
	nameStart := strings.Index(input, "\">") + 2
	nameEnd:= strings.Index(input, "</a>")

	runes := []rune(input)
	countryName := string(runes[nameStart:nameEnd])
	log.Println(countryName)

	input = input[nameEnd+8:]
	countryCode, codeLastIndex := extractTagsData(input)
	log.Println(countryCode)

	input = input[codeLastIndex+5:]
	countryLettersRegexp := "<td>([A-Za-z/ ])+</td>"
	_, lettersLastIndex := extractWithRegexp(input, countryLettersRegexp)

	input = input[lettersLastIndex:]
	countryPopulation, popLastIndex := extractTagsData(input)
	log.Println(countryPopulation)

	input = input[popLastIndex+5:]
	areaReg := "<td>([0-9,])+</td>"
	_, areaLastIndex := extractWithRegexp(input, areaReg)

	input = input[areaLastIndex:]
	countryGdp, _ := extractTagsData(input)
	log.Println(countryGdp)

	//myCountry.Name = countryName
	//myCountry.Code = countryCode
	//err := writeToFile(myCountry)
	//if err != nil {
	//	return err
	//}
	//
	//return parseTable(string(runes[finishIndex:]))
}

func extractTagsData(input string) (string, int) {
	dataStart := strings.Index(input, "<td>") + 4
	dataEnd := strings.Index(input, "</td>")
	data := input[dataStart:dataEnd]
	return data, dataEnd
}

func extractWithRegexp(input string, reg string)(string,int) {
	dataReg, _ := regexp.Compile(reg)
	data := string(dataReg.Find([]byte(input)))
	dataLastIndex := strings.Index(input, data) + len(data)
	return data, dataLastIndex
}

func writeToFile(data Country) error {
	file, err := os.OpenFile(
		//"countryCodes_asOf_" + time.Stamp + ".txt",
		"countryCodes.txt",
		os.O_APPEND|os.O_WRONLY|os.O_CREATE,
		0666,
	)
	if err != nil {
		return err
	}
	defer file.Close()

	JSONData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	w := bufio.NewWriter(file)
	fmt.Fprintln(w, string(JSONData))
	return w.Flush()
}
