package slidingwindow

import (
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/rcrowley/go-metrics"
)

type clock0 struct{}

func (clock0) Now() time.Time {
	return time.Unix(0, 0)
}

type clock2 struct{}

func (clock2) Now() time.Time {
	return time.Unix(2, 0)
}

var (
	t0 = time.Unix(0, 0)
	t1 = time.Unix(1, 0)
	t2 = time.Unix(2, 0)
)

func TestSlidingWindowSample_Clear(t *testing.T) {
	type fields struct {
		values        []sample
		reservoirSize int
		c             clock
	}
	tests := []struct {
		name   string
		fields fields
	}{
		{
			name: "clears an empty values slice",
			fields: fields{
				values:        []sample{},
				reservoirSize: 3,
				c:             clock0{},
			},
		},
		{
			name: "clears a non-empty values slice",
			fields: fields{
				values:        []sample{{v: 1, t: time.Now()}},
				reservoirSize: 1,
				c:             clock0{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Sample{
				values:        tt.fields.values,
				reservoirSize: tt.fields.reservoirSize,
				mu:            sync.Mutex{},
				c:             tt.fields.c,
			}
			s.Clear()

			if len(s.values) != 0 {
				t.Errorf("sample values are not empty")
			}
		})
	}
}

func TestSlidingWindowSample_Size(t *testing.T) {
	type fields struct {
		values        []sample
		reservoirSize int
		c             clock
	}
	tests := []struct {
		name   string
		fields fields
		want   int
	}{
		{
			name: "counts an empty values slice",
			fields: fields{
				values:        []sample{},
				reservoirSize: 0,
				c:             clock0{},
			},
			want: 0,
		},
		{
			name: "counts a non-empty values slice",
			fields: fields{
				values:        []sample{{v: 1, t: t0}},
				reservoirSize: 1,
				c:             clock0{},
			},
			want: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Sample{
				values:        tt.fields.values,
				reservoirSize: tt.fields.reservoirSize,
				mu:            sync.Mutex{},
				c:             tt.fields.c,
			}
			if got := s.Size(); got != tt.want {
				t.Errorf("Size() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSlidingWindowSample_Snapshot(t *testing.T) {
	type fields struct {
		values        []sample
		reservoirSize int
		c             clock
	}
	tests := []struct {
		name   string
		fields fields
		want   metrics.Sample
	}{
		{
			name: "snapshot an empty list of  values",
			fields: fields{
				values:        []sample{},
				reservoirSize: 3,
				c:             clock0{},
			},
			want: metrics.NewSampleSnapshot(0, []int64{}),
		},
		{
			name: "snapshot a non-empty list of  values",
			fields: fields{
				values:        []sample{{v: 1, t: t0}, {v: 2, t: t0}},
				reservoirSize: 2,
				c:             clock0{},
			},
			want: metrics.NewSampleSnapshot(2, []int64{1, 2}),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Sample{
				values:        tt.fields.values,
				reservoirSize: tt.fields.reservoirSize,
				mu:            sync.Mutex{},
				c:             tt.fields.c,
			}
			if got := s.Snapshot(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Snapshot() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSlidingWindowSample_Update(t *testing.T) {
	type fields struct {
		values        []sample
		reservoirSize int
		c             clock
	}
	type args struct {
		v int64
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   []int64
	}{
		{
			name: "updates an empty list of values",
			fields: fields{
				values:        []sample{},
				reservoirSize: 3,
				c:             clock0{},
			},
			args: args{v: 1},
			want: []int64{1},
		},
		{
			name: "updates a non-empty list of values",
			fields: fields{
				values:        []sample{{v: 1, t: t0}, {v: 2, t: t0}, {v: 3, t: t0}},
				reservoirSize: 3,
				c:             clock0{},
			},
			args: args{v: 3},
			want: []int64{1, 2, 3},
		},
		{
			name: "updates a non-empty list of values that have exceeded the reservoir size",
			fields: fields{
				values:        []sample{{v: 1, t: t0}, {v: 2, t: t0}, {v: 3, t: t0}},
				reservoirSize: 3,
				c:             clock0{},
			},
			args: args{v: 4},
			want: []int64{1, 2, 3},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Sample{
				values:        tt.fields.values,
				reservoirSize: tt.fields.reservoirSize,
				mu:            sync.Mutex{},
				c:             tt.fields.c,
			}
			s.Update(tt.args.v)

			if !reflect.DeepEqual(s.Values(), tt.want) {
				t.Errorf("Update() values = %v, want %v", s.Values(), tt.want)
			}
		})
	}
}

func TestSlidingWindowSample_Values(t *testing.T) {
	type fields struct {
		values        []sample
		reservoirSize int
		c             clock
	}
	tests := []struct {
		name   string
		fields fields
		want   []int64
	}{
		{
			name: "returns an empty list of values",
			fields: fields{
				values:        []sample{},
				reservoirSize: 3,
				c:             clock0{},
			},
			want: []int64{},
		},
		{
			name: "returns a non-empty list of values",
			fields: fields{
				values:        []sample{{v: 1, t: t0}, {v: 2, t: t0}, {v: 3, t: t0}},
				reservoirSize: 3,
				c:             clock0{},
			},
			want: []int64{1, 2, 3},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Sample{
				values:        tt.fields.values,
				reservoirSize: tt.fields.reservoirSize,
				mu:            sync.Mutex{},
				c:             tt.fields.c,
			}
			if got := s.Values(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Values() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSlidingWindowSample_slideWindow(t *testing.T) {
	type fields struct {
		values        []sample
		reservoirSize int
		window        time.Duration
		c             clock
	}

	tests := []struct {
		name   string
		fields fields
		want   []int64
	}{
		{
			name: "doesn't slide an empty list of values",
			fields: fields{
				values:        []sample{},
				reservoirSize: 0,
				window:        time.Second,
				c:             clock0{},
			},
			want: []int64{},
		},
		{
			name: "doesn't slide a non-empty list of values where the first value is less than the window",
			fields: fields{
				values:        []sample{{v: 1, t: t0}},
				reservoirSize: 0,
				window:        time.Second,
				c:             clock0{},
			},
			want: []int64{1},
		},
		{
			name: "slides a non-empty list of values where some of the values are preserved",
			fields: fields{
				values:        []sample{{v: 1, t: t0}, {v: 2, t: t1}, {v: 3, t: t2}},
				reservoirSize: 3,
				window:        time.Second,
				c:             clock2{},
			},
			want: []int64{2, 3},
		},
		{
			name: "slides a non-empty list of values where none of the values are preserved",
			fields: fields{
				values:        []sample{{v: 1, t: t0}, {v: 2, t: t0}, {v: 3, t: t0}},
				reservoirSize: 0,
				window:        time.Second,
				c:             clock2{},
			},
			want: []int64{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Sample{
				values:        tt.fields.values,
				reservoirSize: tt.fields.reservoirSize,
				mu:            sync.Mutex{},
				window:        tt.fields.window,
				c:             tt.fields.c,
			}
			s.slideWindow()
			if got := s.Values(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("slideWindow() = %v, want %v", got, tt.want)
			}
		})
	}
}
