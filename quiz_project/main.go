package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"
)

func main() {
	csvFile := flag.String("csv", "problems.csv", "Consists a csv file in the format of 'question,answer'")
	timeLimit := flag.Int("limit", 30, "Time limit for the quiz in seconds")
	shuffle := flag.Bool("shuffle", false, "Shuffle the questions")
	flag.Parse()

	file, err := os.Open(*csvFile)
	if err != nil {
		exit(fmt.Sprintf("Failed to open the csv file: %s\n", *csvFile))
	}

	r := csv.NewReader(file)
	lines, err := r.ReadAll()
	if err != nil {
		exit("Failed to parse the csv file")
	}
	problems := parseLine(lines)
	if *shuffle {
		shuffleProblems(problems)
	}
	timer := time.NewTimer(time.Duration(*timeLimit) * time.Second)

	correct := 0

	for i, p := range problems {
		fmt.Printf("Problem #%d: %s = ", i+1, p.q)
		answerCh := make(chan string)
		// making this because scanf is blocking
		// so we need to run it in a goroutine
		go func() {
			var answer string
			fmt.Scanf("%s\n", &answer)
			answerCh <- answer
		}()
		select {
		case <-timer.C:
			fmt.Println("\nTime's up!")
			fmt.Printf("You scored %d out of %d\n", correct, len(problems))
			return
		case answer := <-answerCh:
			if answer == p.a {
				correct++
			}
		}
	}
}

func parseLine(lines [][]string) []problem {
	problems := make([]problem, len(lines))

	for i, line := range lines {
		problems[i] = problem{
			q: line[0],
			a: strings.TrimSpace(line[1]),
		}
	}
	return problems
}

type problem struct {
	q string
	a string
}

func exit(msg string) {
	fmt.Println(msg)
	os.Exit(1)
}

func shuffleProblems(problems []problem) {
	for i := range problems {
		randomIndex := rand.Intn(len(problems))
		problems[i], problems[randomIndex] = problems[randomIndex], problems[i]
	}
}
