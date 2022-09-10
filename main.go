package main

import (
	"bufio"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	gd "github.com/bolknote/go-gd"
)

type Entry struct {
	dep     string
	semaine string
	cage    string
	clage   string
	pop     int
	p       int
	t       int
	ti      int
	tr      int
	td      int
}

// check if string is present in slice
func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}

	return false
}

func StrToInt(str string) int {
	// Convert String to Int
	str = strings.Replace(str, ",", ".", -1)
	nonFractionalPart := strings.Split(str, ".")
	real, _ := strconv.ParseFloat(nonFractionalPart[0], 64)
	value := int(math.Round(real))
	return value
}

func downloadFile(filepath string, url string) (err error) {
	// download URL to File

	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	return nil
}

func Filemtime(filename string) (int64, error) {
	// Get last modification time of a file

	fd, err := os.Open(filename)
	if err != nil {
		return 0, err
	}
	defer fd.Close()
	fileinfo, err := fd.Stat()
	if err != nil {
		return 0, err
	}
	return fileinfo.ModTime().Unix(), nil
}

// Convert week number to start and end date
func getStartAndEndDate(week int, year int) (week_start string, week_end string) {
	date := time.Date(year, 0, 0, 0, 0, 0, 0, time.UTC)
	isoYear, isoWeek := date.ISOWeek()
	for date.Weekday() != time.Monday { // iterate back to Monday
		date = date.AddDate(0, 0, -1)
		isoYear, isoWeek = date.ISOWeek()
	}
	for isoYear < year { // iterate forward to the first day of the first week
		date = date.AddDate(0, 0, 1)
		isoYear, isoWeek = date.ISOWeek()
	}
	for isoWeek < week { // iterate forward to the first day of the given week
		date = date.AddDate(0, 0, 1)
		isoYear, isoWeek = date.ISOWeek()
	}
	dateend := date.AddDate(0, 0, 5)
	datestart := date.AddDate(0, 0, -1)
	week_start = datestart.Format("02") + "/" + datestart.Format("01")
	week_end = dateend.Format("02") + "/" + dateend.Format("01")
	return
}

// return incidence for a given week and age class
func getIncidence(semaine string, annee string, age string, dept string, dataptr *[]Entry) int {
	pourcentmille := 0
	valeursemaine := ""
	if StrToInt(semaine) > 9 {
		valeursemaine = annee + "-S" + semaine
	} else {
		valeursemaine = annee + "-S0" + semaine
	}
	if len(dept) == 1 {
		dept = "0" + dept
	}
	for _, ele := range *dataptr {
		if (ele.semaine == valeursemaine) && (ele.clage == age) && (ele.dep == dept) {
			if ele.pop != 0 {
				pourcentmille = int(ele.p * 100000 / ele.pop)
			} else {
				pourcentmille = 0
			}
			break
		}
	}
	return pourcentmille
}

// return total incidence for a given week and age class
func getIncidenceTotale(semaine string, annee string, dept string, dataptr *[]Entry) int {
	pourcentmille := 0
	valeursemaine := ""
	if StrToInt(semaine) > 9 {
		valeursemaine = annee + "-S" + semaine
	} else {
		valeursemaine = annee + "-S0" + semaine
	}
	if len(dept) == 1 {
		dept = "0" + dept
	}
	totalvalue := 0
	totalpopulation := 0
	for _, ele := range *dataptr {
		if (ele.semaine == valeursemaine) && (ele.dep == dept) {
			totalvalue = totalvalue + ele.p
			totalpopulation = totalpopulation + ele.pop
		}
	}
	if totalpopulation != 0 {
		pourcentmille = int(totalvalue * 100000 / totalpopulation)
	} else {
		pourcentmille = 0
	}
	return pourcentmille
}

// return last (week,year) filled in data
func getLastData(dataptr *[]Entry) (week int, year int) {
	var slicedate []string
	for _, ele := range *dataptr {
		if !contains(slicedate, ele.semaine) {
			slicedate = append(slicedate, ele.semaine)
		}
	}
	sort.Strings(slicedate)
	lastdate := slicedate[len(slicedate)-1]
	year = StrToInt(lastdate[0:4])
	week = StrToInt(lastdate[6:8])
	return
}

// Draw box
func dessineBox(imptr *gd.Image, posx int, posy int, incidence int) {
	var red, green, blue int
	indexincidence := 0
	im := *imptr
	if incidence < 800 {
		indexincidence = incidence
	} else {
		indexincidence = 800
	}

	gety := 791 - int(indexincidence*560/800)
	rgb := im.ColorAt(50, gety)

	red = int((rgb >> 16) & 0xFF)
	green = int((rgb >> 8) & 0xFF)
	blue = int(rgb & 0xFF)

	im.FilledRectangle(posx, posy, posx+64, posy+78, im.ColorAllocate(red, green, blue))

	texte := strconv.Itoa(incidence)
	txtx := int(posx + (64-len(texte)*11)/2)

	// Download font at https://www.dafont.com/share-techmono.font
	font := "./ShareTechMono-Regular.ttf"

	white := im.ColorAllocate(255, 255, 255)

	im.StringFT(white, font, 14, 0, txtx, posy+47, texte)

}

func createHeatmap(dept string, dataptr *[]Entry) {
	var startweek, startyear int

	nomfichierjpg := "heatmaps_deps/heatmap_taux_" + dept + ".jpeg"
	im := gd.CreateTrueColor(2600, 1100)
	background_color := im.ColorAllocate(255, 255, 255)
	chaine := ""

	im.Fill(0, 0, background_color)

	// Create gradient
	gradient := gd.CreateFromPng("gradient.png")

	gradient.Copy(im, 45, 230, 0, 0, 23, 560)

	valeur := 800
	posx := 70
	posy := 237
	// Font can be downloaded at https://fr.fonts2u.com/open-sans-condensed-light.police
	font := "./OpenSansCondensed-Light.ttf"

	black := im.ColorAllocate(0, 0, 0)
	// Gradient scale
	for i := 1; i <= 5; i++ {
		chaine = "- " + strconv.Itoa(valeur) + " cas"
		im.StringFT(black, font, 14, 0, posx, posy, chaine)

		posy = posy + 139
		valeur = valeur - 200
	}

	// Title

	titre := "Taux d'incidence du Covid19 en fonction de l'âge - Dép. " + dept

	posy = 55
	posx = int((2600 - len(titre)*12) / 2)

	im.StringFT(black, font, 24, 0, posx, posy, titre)

	// Get last week from data

	currentWeekNumber, currentYear := getLastData(dataptr)

	if currentWeekNumber > 31 {
		startweek = currentWeekNumber - 31
		startyear = currentYear
	} else {
		startyear = currentYear - 1
		startweek = 52 - (32 - currentWeekNumber)
	}

	// Construct abscissa
	posx = 360
	posy = 960

	crtweek := startweek
	crtyear := startyear
	for i := 1; i <= 32; i++ {

		debut, fin := getStartAndEndDate(crtweek, crtyear)

		im.StringFT(black, font, 10, 0, posx, posy, debut)
		im.StringFT(black, font, 10, 0, posx, posy+14, fin)

		crtweek = crtweek + 1
		if crtweek == 53 {
			crtweek = 1
			crtyear = crtyear + 1
		}
		posx = posx + 64
	}

	// Draw "0-9" years
	posx = 340
	posy = 787
	crtweek = startweek
	crtyear = startyear
	for i := 1; i <= 32; i++ {
		incidence := getIncidence(strconv.Itoa(crtweek), strconv.Itoa(crtyear), "09", dept, dataptr)
		dessineBox(im, posx, posy, incidence)
		crtweek = crtweek + 1
		if crtweek == 53 {
			crtweek = 1
			crtyear = crtyear + 1
		}
		posx = posx + 64

	}
	im.StringFT(black, font, 16, 0, posx+20, posy+50, "0 à 9 ans")

	// Draw "10-19" years
	posx = 340
	posy = posy - 78
	crtweek = startweek
	crtyear = startyear
	for i := 1; i <= 32; i++ {
		incidence := getIncidence(strconv.Itoa(crtweek), strconv.Itoa(crtyear), "19", dept, dataptr)
		dessineBox(im, posx, posy, incidence)
		crtweek = crtweek + 1
		if crtweek == 53 {
			crtweek = 1
			crtyear = crtyear + 1
		}
		posx = posx + 64

	}
	im.StringFT(black, font, 16, 0, posx+20, posy+50, "10 à 19 ans")

	// Draw "20-29" years
	posx = 340
	posy = posy - 78
	crtweek = startweek
	crtyear = startyear
	for i := 1; i <= 32; i++ {
		incidence := getIncidence(strconv.Itoa(crtweek), strconv.Itoa(crtyear), "29", dept, dataptr)
		dessineBox(im, posx, posy, incidence)
		crtweek = crtweek + 1
		if crtweek == 53 {
			crtweek = 1
			crtyear = crtyear + 1
		}
		posx = posx + 64

	}
	im.StringFT(black, font, 16, 0, posx+20, posy+50, "20 à 29 ans")

	// Draw "30-39" years
	posx = 340
	posy = posy - 78
	crtweek = startweek
	crtyear = startyear
	for i := 1; i <= 32; i++ {
		incidence := getIncidence(strconv.Itoa(crtweek), strconv.Itoa(crtyear), "39", dept, dataptr)
		dessineBox(im, posx, posy, incidence)
		crtweek = crtweek + 1
		if crtweek == 53 {
			crtweek = 1
			crtyear = crtyear + 1
		}
		posx = posx + 64

	}

	im.StringFT(black, font, 16, 0, posx+20, posy+50, "30 à 39 ans")

	// Draw "40-49" years
	posx = 340
	posy = posy - 78
	crtweek = startweek
	crtyear = startyear
	for i := 1; i <= 32; i++ {
		incidence := getIncidence(strconv.Itoa(crtweek), strconv.Itoa(crtyear), "49", dept, dataptr)
		dessineBox(im, posx, posy, incidence)
		crtweek = crtweek + 1
		if crtweek == 53 {
			crtweek = 1
			crtyear = crtyear + 1
		}
		posx = posx + 64

	}
	im.StringFT(black, font, 16, 0, posx+20, posy+50, "40 à 49 ans")

	// Draw "50-59" years
	posx = 340
	posy = posy - 78
	crtweek = startweek
	crtyear = startyear
	for i := 1; i <= 32; i++ {
		incidence := getIncidence(strconv.Itoa(crtweek), strconv.Itoa(crtyear), "59", dept, dataptr)
		dessineBox(im, posx, posy, incidence)
		crtweek = crtweek + 1
		if crtweek == 53 {
			crtweek = 1
			crtyear = crtyear + 1
		}
		posx = posx + 64

	}
	im.StringFT(black, font, 16, 0, posx+20, posy+50, "50 à 59 ans")

	// Draw "60-69" years
	posx = 340
	posy = posy - 78
	crtweek = startweek
	crtyear = startyear
	for i := 1; i <= 32; i++ {
		incidence := getIncidence(strconv.Itoa(crtweek), strconv.Itoa(crtyear), "69", dept, dataptr)
		dessineBox(im, posx, posy, incidence)
		crtweek = crtweek + 1
		if crtweek == 53 {
			crtweek = 1
			crtyear = crtyear + 1
		}
		posx = posx + 64

	}
	im.StringFT(black, font, 16, 0, posx+20, posy+50, "60 à 69 ans")

	// Draw "70-79" years
	posx = 340
	posy = posy - 78
	crtweek = startweek
	crtyear = startyear
	for i := 1; i <= 32; i++ {
		incidence := getIncidence(strconv.Itoa(crtweek), strconv.Itoa(crtyear), "79", dept, dataptr)
		dessineBox(im, posx, posy, incidence)
		crtweek = crtweek + 1
		if crtweek == 53 {
			crtweek = 1
			crtyear = crtyear + 1
		}
		posx = posx + 64

	}
	im.StringFT(black, font, 16, 0, posx+20, posy+50, "70 à 79 ans")

	// Draw "80-89" years
	posx = 340
	posy = posy - 78
	crtweek = startweek
	crtyear = startyear
	for i := 1; i <= 32; i++ {
		incidence := getIncidence(strconv.Itoa(crtweek), strconv.Itoa(crtyear), "89", dept, dataptr)
		dessineBox(im, posx, posy, incidence)
		crtweek = crtweek + 1
		if crtweek == 53 {
			crtweek = 1
			crtyear = crtyear + 1
		}
		posx = posx + 64

	}
	im.StringFT(black, font, 16, 0, posx+20, posy+50, "80 à 89 ans")

	// Draw "+90" years
	posx = 340
	posy = posy - 78
	crtweek = startweek
	crtyear = startyear
	for i := 1; i <= 32; i++ {
		incidence := getIncidence(strconv.Itoa(crtweek), strconv.Itoa(crtyear), "90", dept, dataptr)
		dessineBox(im, posx, posy, incidence)
		crtweek = crtweek + 1
		if crtweek == 53 {
			crtweek = 1
			crtyear = crtyear + 1
		}
		posx = posx + 64

	}
	im.StringFT(black, font, 16, 0, posx+20, posy+50, "+ 90 ans")

	// Draw all ages
	posx = 340
	posy = 865
	crtweek = startweek
	crtyear = startyear
	for i := 1; i <= 32; i++ {
		incidence := getIncidenceTotale(strconv.Itoa(crtweek), strconv.Itoa(crtyear), dept, dataptr)
		dessineBox(im, posx, posy, incidence)
		crtweek = crtweek + 1
		if crtweek == 53 {
			crtweek = 1
			crtyear = crtyear + 1
		}
		posx = posx + 64

	}
	im.StringFT(black, font, 16, 0, posx+20, posy+50, "+ 90 ans")

	im.Jpeg(nomfichierjpg, 95)
}

func main() {

	var data []Entry

	// Download datafile if previous download > 1800s

	url := "https://www.data.gouv.fr/fr/datasets/r/61c53cb2-a400-40b6-81be-59d67266181f"
	datafile := "data.csv"

	if _, err := os.Stat(datafile); err == nil {
		stamp, err := Filemtime(datafile)
		if time.Now().Unix()-stamp > 1800 {
			if downloadFile(datafile, url) != nil {
				fmt.Println("Error when downloading file !")
				fmt.Printf("Error : %s \n", err.Error())
				os.Exit(1)
			}
		}
	} else {
		if downloadFile(datafile, url) != nil {
			fmt.Println("Error when downloading file !")
			fmt.Printf("Error : %s \n", err.Error())
			os.Exit(1)
		}
	}

	// open file

	f, err := os.Open(datafile)
	if err != nil {
		fmt.Println("Error when opening datafile !")
		fmt.Printf("Error : %s \n", err.Error())
		os.Exit(1)
	}

	// close the file at the end of the program
	defer f.Close()

	// read csv file line by line
	scanner := bufio.NewScanner(f)
	skipfirstline := 0
	for scanner.Scan() {
		line := scanner.Text()
		if skipfirstline == 1 {
			rec := strings.Split(line, ";")
			linesize := len(rec)

			if linesize > 9 {
				// Upload datas in slice
				entree := Entry{rec[0], rec[1], rec[2], rec[3], StrToInt(rec[4]), StrToInt(rec[5]), StrToInt(rec[6]), StrToInt(rec[7]), StrToInt(rec[8]), StrToInt(rec[9])}
				data = append(data, entree)
			}
		} else {
			skipfirstline = 1
		}
	}

	// create heatmaps
	departements := []string{"01", "02", "03", "04", "05", "06", "07", "08", "09", "10", "11", "12", "13", "14", "15", "16", "17", "18", "19",
		"21", "22", "23", "24", "25", "26", "27", "28", "29", "30", "31", "32", "33", "34", "35", "36", "37", "38", "39",
		"40", "41", "42", "43", "44", "45", "46", "47", "48", "49", "50", "51", "52", "53", "54", "55", "56", "57", "58", "59",
		"60", "61", "62", "63", "64", "65", "66", "67", "68", "69", "70", "71", "72", "73", "74", "75", "76", "77", "78", "79",
		"80", "81", "82", "83", "84", "85", "86", "87", "88", "89", "90", "91", "92", "93", "94", "95"}

	for _, selecteddpt := range departements {
		fmt.Printf("Creating heatmap for departement %s \n", selecteddpt)
		createHeatmap(selecteddpt, &data)
	}
}
