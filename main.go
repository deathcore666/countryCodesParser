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
	"strconv"
)

type Country struct {
	Name string `json:"name"`
	Code []string `json:"code"`
	GdpPercapita  float64  `json:"GdpPerCapita"`
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

	err = parseTable(resultString)
	if err != nil {
		panic(err)
	}
}

func parseTable(input string) error {
	finishIndex := strings.Index(input, "</tr>")
	if finishIndex == -1 {
		return nil
	}
	finishIndex += 4

	var myCountry Country
	nameStart := strings.Index(input, "\">") + 2
	nameEnd := strings.Index(input, "</a>")
	if nameEnd == -1 {
		return nil
	}
	countryName := input[nameStart:nameEnd]
	log.Println(countryName)

	input = input[nameEnd+8:]
	countryCode, codeLastIndex := extractTagsData(input)
	countryCodes := strings.Split(countryCode, ", ")
	log.Println(countryCode)

	input = input[codeLastIndex+5:]
	countryLettersRegexp := "<td>([A-Za-z/ ])+</td>"
	_, lettersLastIndex := extractWithRegexp(input, countryLettersRegexp)

	input = input[lettersLastIndex:]
	countryPopulation, popLastIndex := extractTagsData(input)
	countryPopulation = strings.Replace(countryPopulation, ",", "", -1)
	popFloat, err := strconv.ParseFloat(countryPopulation, 64)
	if err != nil {
		log.Println(err)
		return err
	}
	log.Println(popFloat)

	input = input[popLastIndex+5:]
	areaReg := "<td>([0-9,])+</td>"
	_, areaLastIndex := extractWithRegexp(input, areaReg)

	input = input[areaLastIndex:]
	countryGdp, _ := extractTagsData(input)

	var multiplier float64
	if strings.Contains(countryGdp, "Billion") {
		multiplier = 1000000000
	}
	if strings.Contains(countryGdp, "Million") {
		multiplier = 1000000
	}
	if strings.Contains(countryGdp, "Trillion") {
		multiplier = 1000000000000
	}

	countryGdp = strings.Replace(countryGdp, " Trillion", "", -1)
	countryGdp = strings.Replace(countryGdp, " Billion", "", -1)
	countryGdp = strings.Replace(countryGdp, " Million", "", -1)

	if countryGdp == "" {
		countryGdp = "0"
	}
	gdpFloat, err := strconv.ParseFloat(countryGdp, 64)
	if err != nil {
		log.Println(err)
		return err
	}
	gdpFloat *= multiplier
	log.Println(strconv.FormatFloat(gdpFloat, 'f', -1, 64))

	myCountry.Name = countryName
	myCountry.Code = countryCodes
	myCountry.GdpPercapita = -1
	if popFloat != 0 {
		myCountry.GdpPercapita = gdpFloat / popFloat
	}

	err = writeToFile(myCountry)
	if err != nil {
		log.Println(err)
		return err
	}

	return parseTable(input)
}

func extractTagsData(input string) (string, int) {
	dataStart := strings.Index(input, "<td>") + 4
	dataEnd := strings.Index(input, "</td>")
	data := input[dataStart:dataEnd]
	input = input[dataEnd+5:]
	return data, dataEnd
}

func extractWithRegexp(input string, reg string) (string, int) {
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
