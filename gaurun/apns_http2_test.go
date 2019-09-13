package gaurun

import (
	"testing"

	"github.com/nohana/buford/push"
	"github.com/stretchr/testify/assert"
)

func TestNewApnsClientHttp2(t *testing.T) {
	t.Run("when send push type none", func(t *testing.T) {
		req := &RequestGaurunNotification{}
		headers := NewApnsHeadersHttp2(req)
		assert.Equal(t, push.PushTypeAlert, headers.PushType)
	})
	t.Run("when send push type 'alert'", func(t *testing.T) {
		req := &RequestGaurunNotification{PushType: string(push.PushTypeAlert)}
		headers := NewApnsHeadersHttp2(req)
		assert.Equal(t, push.PushTypeAlert, headers.PushType)
	})
	t.Run("when send push type 'background'", func(t *testing.T) {
		req := &RequestGaurunNotification{PushType: string(push.PushTypeBackground)}
		headers := NewApnsHeadersHttp2(req)
		assert.Equal(t, push.PushTypeBackground, headers.PushType)
	})
}
