// go port of https://github.com/github/linguist/blob/master/lib/linguist/tokenizer.rb
//
// in their words:
//
//  # Generic programming language tokenizer.
//  #
//  # Tokens are designed for use in the language bayes classifier.
//  # It strips any data strings or comments and preserves significant
//  # language symbols.
//
package tokenizer

import (
	"bufio"
	"bytes"
	"regexp"
)

var (
	// Maximum input length for Tokenize()
	ByteLimit = 100000

	// NOTE(tso): these string slices are turned into their regexp slice counterparts
	// by this package's init() function.
	StartLineComments = []string{
		"\"", // Vim
		"%",  // Tex
	}
	SingleLineComments = []string{
		"//", // C
		"--", // Ada, Haskell, AppleScript
		"#",  // Perl, Bash, Ruby
	}
	MultiLineComments = [][]string{
		{"/*", "*/"},    // C
		{"<!--", "-->"}, // XML
		{"{-", "-}"},    // Haskell
		{"(*", "*)"},    // Coq
		{`"""`, `"""`},  // Python
		{"'''", "'''"},  // Python
		{"#`(", ")"},    // Perl6
	}
	StartLineComment       []*regexp.Regexp
	BeginSingleLineComment []*regexp.Regexp
	BeginMultiLineComment  []*regexp.Regexp
	EndMultiLineComment    []*regexp.Regexp
	String                 = regexp.MustCompile(`[^\\]*(["'` + "`])")
	Shebang                = regexp.MustCompile(`#!.*$`)
	Number                 = regexp.MustCompile(`(0x[0-9a-f]([0-9a-f]|\.)*|\d(\d|\.)*)([uU][lL]{0,2}|([eE][-+]\d*)?[fFlL]*)`)
)

func init() {
	for _, st := range append(StartLineComments, SingleLineComments...) {
		StartLineComment = append(StartLineComment, regexp.MustCompile(`^\s*`+regexp.QuoteMeta(st)))
	}
	for _, sl := range SingleLineComments {
		BeginSingleLineComment = append(BeginSingleLineComment, regexp.MustCompile(regexp.QuoteMeta(sl)))
	}
	for _, ml := range MultiLineComments {
		BeginMultiLineComment = append(BeginMultiLineComment, regexp.MustCompile(regexp.QuoteMeta(ml[0])))
		EndMultiLineComment = append(EndMultiLineComment, regexp.MustCompile(regexp.QuoteMeta(ml[1])))
	}
}

// If the given token matches the start of a multi-line comment,
// this function will return true and a regex for the corresponding closing token,
// otherwise false and nil.
func FindMultiLineComment(token []byte) (matched bool, terminator *regexp.Regexp) {
	for idx, re := range BeginMultiLineComment {
		if re.Match(token) {
			return true, EndMultiLineComment[idx]
		}
	}
	return false, nil
}

// Simple tokenizer that uses bufio.Scanner to process lines and individual words
// and matches them against regular expressions to filter out comments, strings, and numerals
// in a manner very similar to github's linguist (see https://github.com/github/linguist/blob/master/lib/linguist/tokenizer.rb)
//
// The intention is to merely retrieve significant tokens from a piece of source code
// in order to identify the programming language using statistical analysis
// and NOT to be used as any part of the process of compilation whatsoever.
//
// NOTE(tso): The tokens produced by this function may be of a dubious quality due to the approach taken.
// Feedback and alternate implementations welcome :)
func Tokenize(input []byte) (tokens []string) {
	if len(input) == 0 {
		return tokens
	}
	if len(input) >= ByteLimit {
		input = input[:ByteLimit]
	}

	var (
		ml_in   = false                // in a multiline comment
		ml_end  *regexp.Regexp         // closing token regexp
		str_in                 = false // in a string literal
		str_end byte           = 0     // closing token byte to be found by the String regexp
	)

	buf := bytes.NewBuffer(input)
	scanlines := bufio.NewScanner(buf)
	scanlines.Split(bufio.ScanLines)

	// NOTE(tso): the use of goto here is probably interchangable with continue
line:
	for scanlines.Scan() {
		ln := scanlines.Bytes()

		for _, re := range StartLineComment {
			if re.Match(ln) {
				goto line
			}
		}

		// NOTE(tso): bufio.Scanner.Split(bufio.ScanWords) seems to just split on whitespace
		// this may yield inaccurate results where there is a lack of sufficient
		// whitespace for the approaches taken here, i.e. jumping straight to the
		// next word/line boundary.
		ln_buf := bytes.NewBuffer(ln)
		scanwords := bufio.NewScanner(ln_buf)
		scanwords.Split(bufio.ScanWords)
	word:
		for scanwords.Scan() {
			tk_b := scanwords.Bytes()
			tk_s := scanwords.Text()

			// find end of multi-line comment
			if ml_in {
				if ml_end.Match(tk_b) {
					ml_in = false
					ml_end = nil
				}
				goto word
			}

			// find end of string literal
			if str_in {
				s := String.FindSubmatch(tk_b)
				if s != nil && s[1][0] == str_end {
					str_in = false
					str_end = 0
				}
				goto word
			}

			// find single-line comment
			for _, re := range BeginSingleLineComment {
				if re.Match(tk_b) {
					goto line
				}
			}

			// find start of multi-line comment
			if matched, terminator := FindMultiLineComment(tk_b); matched {
				ml_in = true
				ml_end = terminator
				goto word
			}

			// find start of string literal
			if s := String.FindSubmatch(tk_b); s != nil {
				str_in = true
				str_end = s[1][0]
				goto word
			}

			// find numeric literal
			if n := Number.Find(tk_b); n != nil {
				goto word
			}

			// add valid tokens to result set
			tokens = append(tokens, tk_s)
		}
	}
	return tokens
}
