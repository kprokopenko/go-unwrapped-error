package p

import (
	"errors"
	"fmt"
	"os"
)

func testMoreThanOneReturn() error {
	if _, err := os.Open("test"); err != nil {
		return err // want "unwrapped error"
	}

	_, err := os.Open("test2")

	return err // want "unwrapped error"
}

func testMoreThanOneReturnAndNil() error {
	if _, err := os.Open("test"); err != nil {
		return err
	}

	return nil
}

func testMoreThanOneReturnsNewError() error {
	if _, err := os.Open("test"); err != nil {
		return err
	}

	return errors.New("abc")
}

func testMoreThanOneReturnNoError() *os.File {
	if f, err := os.Open("test"); err == nil {
		return f
	}

	f, _ := os.Open("test2")

	return f
}

func testOnlyOneReturn() error {
	_, err := os.Open("test2")

	return err
}

func testWrappedError() error {
	if _, err := os.Open("test"); err != nil {
		return fmt.Errorf("open test file: %w", err)
	}

	if _, err := os.Open("test2"); err != nil {
		return fmt.Errorf("open test2 file: %w", err)
	}

	return nil
}

func testNakedSecondFunctionReturn() (*os.File, error) {
	if _, err := os.Open("test"); err != nil {
		return nil, fmt.Errorf("open test file: %w", err)
	}

	return os.Open("test2")
}

func testNoReturn() {}

func testNoErrorReturn() string { return "" }
