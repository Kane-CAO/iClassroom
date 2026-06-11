package domain

import "testing"

func TestRoomStatusValid(t *testing.T) {
	valid := []RoomStatus{RoomStatusCreated, RoomStatusActive, RoomStatusEnded}
	for _, s := range valid {
		if !s.Valid() {
			t.Errorf("RoomStatus(%q).Valid() = false, want true", s)
		}
	}
	if RoomStatus("archived").Valid() {
		t.Error(`RoomStatus("archived").Valid() = true, want false`)
	}
}

func TestTaskStatusValid(t *testing.T) {
	valid := []TaskStatus{TaskStatusPublished, TaskStatusPaused, TaskStatusClosed}
	for _, s := range valid {
		if !s.Valid() {
			t.Errorf("TaskStatus(%q).Valid() = false, want true", s)
		}
	}
	if TaskStatus("draft").Valid() {
		t.Error(`TaskStatus("draft").Valid() = true, want false`)
	}
}

func TestTargetTypeValid(t *testing.T) {
	if !TargetAll.Valid() || !TargetGroups.Valid() {
		t.Error("TargetAll/TargetGroups should be valid")
	}
	if TargetType("everyone").Valid() {
		t.Error(`TargetType("everyone").Valid() = true, want false`)
	}
}

func TestSubmissionStatusValid(t *testing.T) {
	if !SubmissionStatusSubmitted.Valid() || !SubmissionStatusGraded.Valid() {
		t.Error("submitted/graded should be valid")
	}
	if SubmissionStatus("pending").Valid() {
		t.Error(`SubmissionStatus("pending").Valid() = true, want false`)
	}
}

func TestDisplayModeValid(t *testing.T) {
	if !DisplayAnonymous.Valid() || !DisplayShowGroup.Valid() {
		t.Error("anonymous/showGroup should be valid")
	}
	if DisplayMode("public").Valid() {
		t.Error(`DisplayMode("public").Valid() = true, want false`)
	}
}

func TestIsValidScore(t *testing.T) {
	cases := []struct {
		score int
		want  bool
	}{
		{0, false},  // 0 is explicitly disallowed
		{1, true},   // lower bound
		{8, true},   // mid
		{10, true},  // upper bound
		{11, false}, // above range
		{-3, false}, // negative
	}
	for _, c := range cases {
		if got := IsValidScore(c.score); got != c.want {
			t.Errorf("IsValidScore(%d) = %v, want %v", c.score, got, c.want)
		}
	}
}
