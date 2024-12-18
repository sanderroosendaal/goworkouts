package goworkouts

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/google/uuid"
	"github.com/tormoder/fit"
)

/*
func TestRowsandallFit(t *testing.T) {
	w, err := ReadFit("testdata/fitsdk/session.fit")
	if err != nil {
		fmt.Println(err)
		t.Errorf("ReadFit returned an error")
	}
	_, err = w.ToJSON()
	if err != nil {
		t.Errorf("Could not convert to JSON")
	}
	// fmt.Println(string(wjson))
    }
*/

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

func TestDecodeJSONOK2(t *testing.T) {
	var wjson string
	wjson = "{\"name\": \"\", \"sport\": \"rowing\", \"filename\": \"/home/sander/python/rowsandall/media/6cf0ad19-3bb6-4620-927d-e963c5b57be5.fit\", \"steps\": [{\"wkt_step_name\": \"0\", \"stepId\": 0, \"durationType\": \"Time\", \"durationValue\": 120000, \"intensity\": \"Active\", \"targetType\": \"Cadence\", \"targetValue\": 22}, {\"wkt_step_name\": \"1\", \"stepId\": 1, \"durationType\": \"Time\", \"durationValue\": 120000, \"intensity\": \"Rest\"}, {\"wkt_step_name\": \"2\", \"stepId\": 2, \"durationType\": \"Time\", \"durationValue\": 120000, \"intensity\": \"Active\", \"targetType\": \"Cadence\", \"targetValue\": 22}, {\"wkt_step_name\": \"3\", \"stepId\": 3, \"durationType\": \"Time\", \"durationValue\": 120000, \"intensity\": \"Rest\"}, {\"wkt_step_name\": \"4\", \"stepId\": 4, \"durationType\": \"Time\", \"durationValue\": 120000, \"intensity\": \"Active\", \"targetType\": \"Cadence\", \"targetValue\": 22}, {\"wkt_step_name\": \"5\", \"stepId\": 5, \"durationType\": \"Time\", \"durationValue\": 120000, \"intensity\": \"Rest\"}, {\"wkt_step_name\": \"6\", \"stepId\": 6, \"durationType\": \"Time\", \"durationValue\": 120000, \"intensity\": \"Active\", \"targetType\": \"Cadence\", \"targetValue\": 22}, {\"wkt_step_name\": \"7\", \"stepId\": 7, \"durationType\": \"Time\", \"durationValue\": 120000, \"intensity\": \"Rest\"}]}"
	_, err := FromJSON(wjson)
	if err != nil {
		fmt.Println(err.Error())
		t.Errorf("Got Error")
	}
}

func TestReadFit3(t *testing.T) {
	w, err := ReadFit("testdata/fitsdk/WorkoutRepeatSteps.fit")
	_, err = w.ToJSON()
	if err != nil {
		t.Errorf("ReadFit returned an error")
	}
	// fmt.Println(string(wjson))
}

func TestReadFit4(t *testing.T) {
	_, err := ReadFit("testdata/fitsdk/WorkoutRepeatGreaterThanStep.fit")
	if err != nil {
		t.Errorf("ReadFit returned an error")
	}
}

func TestReadFit5(t *testing.T) {
	w, err := ReadFit("testdata/rowingworkout.fit")
	if err != nil {
		t.Errorf("ReadFit returned an error")
	}
	// test if w.Sport is set correctly
	if w.Sport != "rowing" {
		t.Errorf("Sport is not set correctly")
	}
	wjson, err := w.ToJSON()
	if err != nil {
		t.Errorf("ReadFit returned an error")
	}
	// check if the json contains a key "sport" with value "rowing"
	if !bytes.Contains(wjson, []byte("\"sport\":\"rowing\"")) {
		t.Errorf("JSON does not contain sport rowing")
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

func TestReadFittoIntervals(t *testing.T) {
	w, err := ReadFit("testdata/rowingworkout.fit")
	if err != nil {
		t.Errorf("ReadFit returned an error")
	}
	_, err = w.ToIntervals()
	if err != nil {
		t.Errorf("ToIntervals returned an error")
	}

}

func TestReadFittoIntervals2(t *testing.T) {
	w, err := ReadFit("testdata/repeats.fit")
	if err != nil {
		t.Errorf("ReadFit returned an error")
	}
	s, err := w.ToIntervals()
	if err != nil {
		t.Errorf("ToIntervals returned an error")
	}
	fmt.Println(s)

}

func TestReadFittoYAML(t *testing.T) {
	w, err := ReadFit("testdata/fitsdk/WorkoutCustomTargetValues.fit")
	if err != nil {
		t.Errorf("ReadFit returned an error")
	}
	wyaml, err := w.ToYAML()
	//fmt.Println(len(wyaml))
	//fmt.Println(string(wyaml))
	if len(wyaml) != 922 {
		t.Errorf("ToYAML returned a string of a different length")
	}
	if err != nil {
		t.Errorf("ToYAML returned an error")
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
	_, err = fit.Decode(bytes.NewReader(data))
	if err != nil {
		t.Errorf("Could not read written file")
	}
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

	workoutFile, err := fitf.Workout()
	if err != nil {
		t.Errorf("Could not retrieve Workout from new file")
	}
	steps := workoutFile.WorkoutSteps

	// fmt.Printf("Got %v steps\n", len(steps))

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

func TestTrainingPlan(t *testing.T) {
	w1, err := ReadFit("testdata/fitsdk/WorkoutIndividualSteps.fit")
	if err != nil {
		t.Errorf("Could not read fit file")
	}
	w2, err := ReadFit("testdata/fitsdk/WorkoutRepeatGreaterThanStep.fit")
	if err != nil {
		t.Errorf("Could not read fit file")
	}
	w3, err := ReadFit("testdata/fitsdk/WorkoutRepeatSteps.fit")
	if err != nil {
		t.Errorf("Could not read fit file")
	}
	day1 := TrainingDay{1, []Workout{w1}}
	day2 := TrainingDay{2, []Workout{w2}}
	day3 := TrainingDay{4, []Workout{w3}}

	listofdays := []TrainingDay{day1, day2, day3}

	plan := TrainingPlan{uuid.New(), "", "Test Plan", listofdays, 4, "Description"}
	planJSON, err := json.MarshalIndent(plan, "", "   ")
	if err != nil {
		t.Errorf("Could not convert training plan to json")
	}
	//	fmt.Println(string(planJSON))
	//	fmt.Println(len(planJSON))
	expected := 7674
	if len(planJSON) != expected {
		t.Errorf("Conversion of the training plan to JSON gave the wrong json length. Expected %v, got %v", expected, len(planJSON))
	}
}
