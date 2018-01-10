package linguist

import (
	"bytes"
	"log"
	"math"

	"github.com/Azure/draft/pkg/linguist/data"
	"github.com/Azure/draft/pkg/linguist/tokenizer"
	"github.com/jbrukh/bayesian"
)

var classifier *bayesian.Classifier
var classifier_initialized bool = false

// Gets the baysian.Classifier which has been trained on programming language
// samples from github.com/github/linguist after running the generator
//
// See also cmd/generate-classifier
func getClassifier() *bayesian.Classifier {
	// NOTE(tso): this could probably go into an init() function instead
	// but this lazy loading approach works, and it's conceivable that the
	// analyse() function might not invoked in an actual runtime anyway
	if !classifier_initialized {
		d, err := data.Asset("classifier")
		if err != nil {
			log.Panicln(err)
		}
		reader := bytes.NewReader(d)
		classifier, err = bayesian.NewClassifierFromReader(reader)
		if err != nil {
			log.Panicln(err)
		}
		classifier_initialized = true
	}
	return classifier
}

// Uses Naive Bayesian Classification on the file contents provided.
//
// Returns the name of a programming language, or the empty string if one could
// not be determined.
//
// It is recommended to use LanguageByContents() instead of this function directly.
//
// Obtain hints from LanguageHints()
//
// NOTE(tso): May yield inaccurate results
func Analyse(contents []byte, hints []string) (language string) {
	document := tokenizer.Tokenize(contents)
	classifier := getClassifier()
	scores, idx, _ := classifier.LogScores(document)

	if len(hints) == 0 {
		return string(classifier.Classes[idx])
	}

	langs := map[string]struct{}{}
	for _, hint := range hints {
		langs[hint] = struct{}{}
	}

	best_score := math.Inf(-1)
	best_answer := ""

	for id, score := range scores {
		answer := string(classifier.Classes[id])
		if _, ok := langs[answer]; ok {
			if score >= best_score {
				best_score = score
				best_answer = answer
			}
		}
	}
	return best_answer
}
