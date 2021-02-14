package goworkouts

import (
	"testing"
)

func TestReadFit(t *testing.T) {
	_, err := ReadFit("testdata/fitsdk/WorkoutCustomTargetValues.fit")
	if err != nil {
		t.Errorf("ReadFit returned an error")
	}
}

func TestReadFit2(t *testing.T) {
	_, err := ReadFit("testdata/fitsdk/WorkoutIndividualSteps.fit")
	if err != nil {
		t.Errorf("ReadFit returned an error")
	}
}

func TestReadFit3(t *testing.T) {
	w, err := ReadFit("testdata/fitsdk/WorkoutRepeatSteps.fit")
	_, err = w.ToJSON()
	if err != nil {
		t.Errorf("ReadFit returned an error")
	}
}

func TestReadFit4(t *testing.T) {
	_, err := ReadFit("testdata/fitsdk/WorkoutRepeatGreaterThanStep.fit")
	if err != nil {
		t.Errorf("ReadFit returned an error")
	}
}

func TestReadFitFalse(t *testing.T) {
	_, err := ReadFit("testdata/fitsdk/Activity.fit")
	if err == nil {
		t.Errorf("Should have thrown an error")
	}
}

func TestReadFittoJSON(t *testing.T) {
	w, err := ReadFit("testdata/fitsdk/WorkoutCustomTargetValues.fit")
	if err != nil {
		t.Errorf("ReadFit returned an error")
	}
	wjson, err := w.ToJSON()
	if len(wjson) != 1011 {
		t.Errorf("ToJSON returned a string of a different length")
	}
	if err != nil {
		t.Errorf("ToJSON returned an error")
	}
}
