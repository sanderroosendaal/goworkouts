package goworkouts

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/tormoder/fit"
)

func TestRowsandallFit(t *testing.T) {
	w, err := ReadFit("testdata/session.fit")
	if err != nil {
		t.Errorf("ReadFit returned an error")
	}
	wjson, err := w.ToJSON()
	if err != nil {
		t.Errorf("Could not convert to JSON")
	}
	fmt.Println(string(wjson))
}

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

func TestDecodeJSON(t *testing.T) {
	var wjson string
	wjson = "{\"name\": \"\", \"sport\": \"rowing\", \"filename\": \"\", \"steps\": [{\"wkt_step_name\": \"0\", \"stepId\": 0, \"durationType\": \"Distance\", \"durationValue\": 1000, \"intensity\": \"active\"}, {\"wkt_step_name\": \"1\", \"stepId\": 1, \"durationType\": \"Distance\", \"durationValue\": 1000, \"intensity\": \"active\"}]}"
	_, err := FromJSON(wjson)
	if err != nil {
		fmt.Println(err.Error())
		t.Errorf("Got error")

	}
}

func TestReadFit3(t *testing.T) {
	w, err := ReadFit("testdata/fitsdk/WorkoutRepeatSteps.fit")
	wjson, err := w.ToJSON()
	if err != nil {
		t.Errorf("ReadFit returned an error")
	}
	fmt.Println(string(wjson))
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
	if len(wjson) != 934 {
		t.Errorf("ToJSON returned a string of a different length")
	}
	if err != nil {
		t.Errorf("ToJSON returned an error")
	}
}

func TestReadFittoFIT(t *testing.T) {
	w, err := ReadFit("testdata/fitsdk/WorkoutCustomTargetValues.fit")
	if err != nil {
		t.Errorf("ReadFit returned an error")
	}
	wjfit, err := w.ToFIT()
	if err != nil {
		t.Errorf("ToFit returned an error")
	}
	ok, err := WriteFit("testdata/new.fit", wjfit, true)
	if err != nil {
		t.Errorf("Error writing file")
	}
	if !ok {
		t.Errorf("Not written")
	}
	data, _ := ioutil.ReadFile("testdata/new.fit")
	fitf, err := fit.Decode(bytes.NewReader(data))
	if err != nil {
		t.Errorf("Could not read written file")
	}
	fmt.Println(fitf.FileId.Type)
	fmt.Println(fitf.FileId.TimeCreated)
	fmt.Println(fitf.FileId.Manufacturer)
}

func TestWriter(t *testing.T) {
	data, _ := ioutil.ReadFile("testdata/fitsdk/WorkoutCustomTargetValues.fit")
	fitf, _ := fit.Decode(bytes.NewReader(data))

	oldWorkout, err := fitf.Workout()
	if err != nil {
		t.Errorf("Unable to parse test file")
	}

	oldSteps := oldWorkout.WorkoutSteps

	ok, err := WriteFit("testdata/new.fit", fitf, true)
	if err != nil {
		t.Errorf("Error writing file")
	}
	if !ok {
		t.Errorf("Not written")
	}
	data, _ = ioutil.ReadFile("testdata/new.fit")
	fitf, err = fit.Decode(bytes.NewReader(data))
	if err != nil {
		t.Errorf("Could not read written file")
	}
	fmt.Println(fitf.FileId.Type)
	fmt.Println(fitf.FileId.TimeCreated)
	fmt.Println(fitf.FileId.Manufacturer)

	workoutFile, err := fitf.Workout()
	if err != nil {
		t.Errorf("Could not retrieve Workout from new file")
	}
	steps := workoutFile.WorkoutSteps

	fmt.Printf("Got %v steps\n", len(steps))

	if len(steps) != len(oldSteps) {
		t.Errorf("Reading back new File got incorrect number of steps. Got %d, wanted %d\n", len(steps), len(oldSteps))
	}
	for i, step := range steps {
		if step.WktStepName != oldSteps[i].WktStepName {
			t.Errorf("Expected %v, got %v", oldSteps[i].WktStepName, step.WktStepName)
		}
		if step.DurationType != oldSteps[i].DurationType {
			t.Errorf("Expected %v, got %v", oldSteps[i].DurationType.String(), step.DurationType.String())
		}
		if step.DurationValue != oldSteps[i].DurationValue {
			t.Errorf("Expected %v, got %v", oldSteps[i].DurationValue, step.DurationValue)
		}
	}
}
