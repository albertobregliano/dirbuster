package dirbuster

import (
	"bufio"
	"errors"
	"io"
	"os"
	"strconv"
)

type buster struct {
	input  io.Reader
	output io.Writer
}

type option func(*buster) error

func WithInput(input io.Reader) option {
	return func(b *buster) error {
		if input == nil {
			return errors.New("nil input reader")
		}
		b.input = input
		return nil
	}
}

func WithOutput(output io.Writer) option {
	return func(b *buster) error {
		if output == nil {
			return errors.New("nil output writer")
		}
		b.output = output
		return nil
	}
}

func NewBuster(opts ...option) (buster, error) {
	b := buster{
		input:  os.Stdin,
		output: os.Stdout,
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
