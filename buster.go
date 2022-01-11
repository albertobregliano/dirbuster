package dirbuster

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
)

type buster struct {
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
		if wordlist == "" {
			return errors.New("wordlist cannot be blank")
		}
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

func WithOutput(output interface{}) option {
	return func(b *buster) error {
		switch o := output.(type) {
		case string:
			if o == "" {
				return nil
			}
			outputFile, err := os.Create(o)
			if err != nil {
				return errors.New("impossible to create file")
			}
			b.output = outputFile
		case io.Writer:
			b.output = o
		case nil:
		}
		return nil
	}
}

func NewBuster(opts ...option) (buster, error) {
	b := buster{
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

func Run(ctx context.Context, baseurl, wordlist string, output interface{}) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	b, err := NewBuster(
		WithBaseurl(baseurl),
		WithWordlist(wordlist),
		WithOutput(output),
	)
	if err != nil {
		return fmt.Errorf("impossible to create buster, error: %v", err)
	}

	return Exists(ctx, b)
}
