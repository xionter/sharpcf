package main

import (
	"context"
	"fmt"
	"github.com/chromedp/chromedp"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type Contest struct {
	name          string
	url           string
	remainingTime string
}

func main() {
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", false),
		chromedp.Flag("enable-automation", false),
		chromedp.Flag("disable-blink-features", "AutomationControlled"),
		chromedp.Flag("user-data-dir", "./chromedata"))

	allocCtx, alocCancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer alocCancel()
	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	url := "https://codeforces.com/group/lwQWQuob0B/contests"

	var contestURL []string
	var contestNames []string
	var contestRemainingTime []string
	err := chromedp.Run(ctx, chromedp.Navigate(url),
		chromedp.WaitReady("tr[data-groupcontestid]"),
		chromedp.Evaluate(`
				Array.from(document.querySelectorAll('tr[data-groupcontestid]'))
				.map(contest => contest.children[0].children[1].href)
					`, &contestURL),

		chromedp.Evaluate(`
				Array.from(document.querySelectorAll('tr[data-groupcontestid]')).
				map(contest => contest.children[0].firstChild.data)	
				`, &contestNames),
		chromedp.Evaluate(`
				Array.from(document.querySelectorAll('tr[data-groupcontestid]'))
				.map(contest => contest.children[3]?.children[2])
				.map(standings => standings?.children[0]?.children[0]?.title)	
				`, &contestRemainingTime))

	if err != nil {
		fmt.Printf("couldn't visit given url: %v and extract contests %v", url, err)
		return
	}
	n := len(contestNames)
	for i := range n {
		fmt.Printf("%v) %v ", i+1, strings.TrimSpace(contestNames[n-i-1]))
		if contestRemainingTime[n-i-1] != "" {
			fmt.Printf("(remaining time: %v)", contestRemainingTime[n-i-1])
		}
		fmt.Println()
	}

	fmt.Printf("Specify contest№ to scrape(1 - %v):\n", n)
	var pick int
	_, err = fmt.Scan(&pick)

	if err != nil {
		fmt.Printf("Plese provide valid number. err: %v", err)
		return
	}
	fmt.Println(contestURL[n-pick])
	contestPath := filepath.Join("..", fmt.Sprintf("contest%d", pick))
	handleContest(filepath.Join("..", contestPath), ctx, contestURL[n-pick])
	os.WriteFile("current_contest.txt", []byte(contestPath), 0644)
}

func handleContest(contestPath string, parent context.Context, contest string) {
	ctx, cancel := chromedp.NewContext(parent)
	defer cancel()
	var problems []string
	err := chromedp.Run(ctx, chromedp.Navigate(contest),
		chromedp.WaitReady(".problems"),
		chromedp.Evaluate(`
							Array.from(document.querySelectorAll(".problems > tbody > tr"))
							.slice(1)
							.map(problem => problem.children[0].children[0].href)
							`, &problems))
	if err != nil {
		fmt.Println("coulndt handle contest", err)
		return
	}
	var wg sync.WaitGroup
	for _, problem := range problems {
		idx := strings.LastIndex(problem, "/")
		name := "problem" + problem[idx+1:]
		path := filepath.Join(contestPath, name)
		wg.Go(func() {
			handleProblem(path, ctx, problem)
		})
	}
	wg.Wait()
}

func handleProblem(path string, parent context.Context, problem string) {
	ctx, cancel := chromedp.NewContext(parent)
	defer cancel()

	var inputs []string
	var outputs []string

	err := chromedp.Run(ctx, chromedp.Navigate(problem),
		chromedp.WaitReady("div.input"),

		chromedp.Evaluate(`
				Array.from(document.querySelectorAll("div.input"))
				.map((title) => title.children[1].innerText)
				`, &inputs),

		chromedp.WaitReady("div.output"),
		chromedp.Evaluate(`
				Array.from(document.querySelectorAll("div.output"))
				.map((title) => title.children[1].innerText)
				`, &outputs))

	if err != nil {
		fmt.Println("coulnd't handle contest IO", err)
		return
	}

	testsPath := filepath.Join(path, "tests")
	err = os.MkdirAll(testsPath, 0755)
	if err != nil {
		fmt.Println("couldn't create a tests directory for problem:", testsPath)
		return
	}
	writeData(filepath.Join(testsPath, "input"), inputs)
	writeData(filepath.Join(testsPath, "output"), outputs)
}

func writeData(path string, data []string) {
	for i := range data {
		f, err := os.Create(fmt.Sprintf("%s%d", path, i+1))
		if err != nil {
			fmt.Printf("couldn't create file %v \n", path)
		}
		f.WriteString(data[i])
		defer f.Close()
	}
}
