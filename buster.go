package dirbuster

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"strconv"
)

type buster struct {
	context  context.Context
	baseurl  string
	wordlist string
	input    io.Reader
	output   io.Writer
}

type option func(*buster) error

func WithBaseurl(baseurl string) option {
	return func(b *buster) error {
		_, err := url.Parse(baseurl)
		if baseurl == "" || err != nil {
			return errors.New("baseurl empty or not valid")
		}
		b.baseurl = baseurl
		return nil
	}
}

func WithWordlist(wordlist string) option {
	return func(b *buster) error {
		_, err := os.Stat(wordlist)
		if os.IsNotExist(err) {
			return errors.New("wordlist not reachable")
		}
		b.wordlist = wordlist
		return nil
	}
}

func WithInput(input io.Reader) option {
	return func(b *buster) error {
		if input == nil {
			return errors.New("nil input reader")
		}
		b.input = input
		return nil
	}
}

func WithOutput(output string) option {
	return func(b *buster) error {
		if output == "" {
			return nil
		}
		outputFile, err := os.Create(output)
		if err != nil {
			return errors.New("impossible to create file")
		}
		b.output = outputFile
		return nil
	}
}

func WithContext(ctx context.Context) option {
	return func(b *buster) error {
		if ctx == nil {
			b.context = context.TODO()
		}
		b.context = ctx
		return nil
	}
}

func NewBuster(opts ...option) (buster, error) {
	b := buster{
		context:  context.TODO(),
		baseurl:  "http://127.0.0.1",
		wordlist: "",
		input:    os.Stdin,
		output:   os.Stdout,
	}
	for _, opt := range opts {
		err := opt(&b)
		if err != nil {
			return buster{}, err
		}
	}
	return b, nil
}

func (b buster) Urls() int {
	lines := 0
	scanner := bufio.NewScanner(b.input)
	for scanner.Scan() {
		lines++
	}
	b.output.Write([]byte(strconv.Itoa(lines)))
	return lines
}

func Urls() int {
	b, err := NewBuster()
	if err != nil {
		panic("internal error")
	}
	return b.Urls()
}

func Run(ctx context.Context, baseurl, wordlist, output string) error {
	b, err := NewBuster(
		WithContext(ctx),
		WithBaseurl(baseurl),
		WithWordlist(wordlist),
		WithOutput(output),
	)
	if err != nil {
		return fmt.Errorf("impossible to create buster, error: %v", err)
	}

	Exists(b.context, b)
	return nil
}
