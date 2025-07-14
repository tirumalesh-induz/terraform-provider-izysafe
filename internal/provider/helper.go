package provider

import (
	"fmt"
	"regexp"
)

func ValidPin(val interface{}, key string) (warns []string, errs []error) {
	pin, ok := val.(string)
	if !ok {
		errs = append(errs, fmt.Errorf("expected pin to be a string"))
		return
	}

	match, err := regexp.MatchString(`^\d{6}$`, pin)
	if err != nil {
		errs = append(errs, fmt.Errorf("error validating pin: %v", err))
		return
	}

	if !match {
		errs = append(errs, fmt.Errorf("%q must be a 6-digit number", key))
	}
	return
}

func ValidateNumericString(val interface{}, key string) (warns []string, errs []error) {
	s, ok := val.(string)
	if !ok {
		errs = append(errs, fmt.Errorf("expected %q to be a string", key))
		return
	}

	match, err := regexp.MatchString(`^\d+$`, s)
	if err != nil {
		errs = append(errs, fmt.Errorf("regex error: %v", err))
		return
	}

	if !match {
		errs = append(errs, fmt.Errorf("%q must contain only numeric digits", key))
	}
	return
}
