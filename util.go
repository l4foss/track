package main

import (
	gosu "github.com/thehowl/go-osuapi"
	"sort"
)

type Scores []gosu.GSScore

/*
* functions to sort scores
 */
func (s Scores) Len() int {
	return len(s)
}

func (s Scores) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s Scores) Less(i, j int) bool {
	return s[i].Score.Score < s[j].Score.Score
}

func ScoreSort(scores []gosu.GSScore) {
	sort.Sort(Scores(scores))
}
