package asset

import "testing"

func TestDefineMimeType(t *testing.T) {
	t.Parallel()

	tests := []struct {
		extension string
		want      mimeType
	}{
		{extension: "png", want: png},
		{extension: "PNG", want: png},
		{extension: "jpg", want: jpeg},
		{extension: "jpeg", want: jpeg},
		{extension: "webp", want: webp},
		{extension: "mp4", want: mp4},
		{extension: "mov", want: mov},
		{extension: "mp3", want: mp3},
		{extension: "wav", want: wav},
		{extension: "m4a", want: m4a},
		{extension: "zip", want: zip},
		{extension: "unknown", want: defaultMimeType},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.extension, func(t *testing.T) {
			t.Parallel()

			got := defineMimeType(tt.extension)
			if got != tt.want {
				t.Fatalf("defineMimeType(%q) = %q, want %q", tt.extension, got, tt.want)
			}
		})
	}
}
