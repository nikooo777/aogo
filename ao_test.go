package aogo

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/liteseed/goar/signer"
	"github.com/liteseed/goar/tag"
	"github.com/stretchr/testify/assert"
)

func NewAOMock(CUURL, MUURL string) *AO {
	return &AO{
		cu: newCU(CUURL),
		mu: newMU(MUURL),
	}
}

func setupMU(t *testing.T, handler http.HandlerFunc) *httptest.Server {
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	return server
}

func setupCU(t *testing.T, handler http.HandlerFunc) *httptest.Server {
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	return server
}

func setupSigner(t *testing.T) *signer.Signer {
	s, err := signer.FromPath("./keys/wallet.json")
	assert.NoError(t, err)
	return s
}

func TestSpawnProcess_AO(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		muServer := setupMU(t, func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodPost, r.Method)
			w.WriteHeader(http.StatusOK)
			_, err := w.Write([]byte(`{"id": "mockProcessID"}`))
			assert.NoError(t, err)
		})

		ao := NewAOMock("", muServer.URL)
		data := []byte("test data")
		tags := []tag.Tag{{Name: "TestTag", Value: "TestValue"}}
		s := setupSigner(t)

		id, err := ao.SpawnProcess("testModule", data, tags, s)
		assert.NoError(t, err)
		assert.Equal(t, "mockProcessID", id)
	})

	t.Run("EmptyDataAndTags", func(t *testing.T) {
		muServer := setupMU(t, func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodPost, r.Method)
			w.WriteHeader(http.StatusOK)
			_, err := w.Write([]byte(`{"id": "mockProcessID"}`))
			assert.NoError(t, err)
		})

		ao := NewAOMock("", muServer.URL)
		s := setupSigner(t)

		id, err := ao.SpawnProcess("testModule", nil, nil, s)
		assert.NoError(t, err)
		assert.Equal(t, "mockProcessID", id)
	})

	t.Run("InvalidSigner", func(t *testing.T) {
		ao := NewAOMock("", "")

		_, err := ao.SpawnProcess("testModule", []byte("testData"), nil, nil)
		assert.Error(t, err)
	})

	t.Run("HTTPErrorResponse", func(t *testing.T) {
		muServer := setupMU(t, func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodPost, r.Method)
			w.WriteHeader(http.StatusInternalServerError)
		})

		ao := NewAOMock("", muServer.URL)
		s := setupSigner(t)

		_, err := ao.SpawnProcess("testModule", []byte("testData"), nil, s)
		assert.Error(t, err)
	})
}

func TestSendMessage_AO(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		muServer := setupMU(t, func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodPost, r.Method)
			w.WriteHeader(http.StatusOK)
			_, err := w.Write([]byte(`{"id": "mockMessageID"}`))
			assert.NoError(t, err)
		})

		ao := NewAOMock("", muServer.URL)
		process := "testProcess"
		data := "testData"
		tags := &[]tag.Tag{{Name: "TestTag", Value: "TestValue"}}
		s := setupSigner(t)

		id, err := ao.SendMessage(process, data, tags, "", s)
		assert.NoError(t, err)
		assert.Equal(t, "mockMessageID", id)
	})

	t.Run("EmptyData", func(t *testing.T) {
		muServer := setupMU(t, func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodPost, r.Method)
			w.WriteHeader(http.StatusOK)
			_, err := w.Write([]byte(`{"id": "mockMessageID"}`))
			assert.NoError(t, err)
		})

		ao := NewAOMock("", muServer.URL)
		s := setupSigner(t)

		id, err := ao.SendMessage("testProcess", "", nil, "", s)
		assert.NoError(t, err)
		assert.Equal(t, "mockMessageID", id)
	})

	t.Run("InvalidSigner", func(t *testing.T) {
		ao := NewAOMock("", "")

		_, err := ao.SendMessage("testProcess", "testData", nil, "", nil)
		assert.Error(t, err)
	})

	t.Run("HTTPErrorResponse", func(t *testing.T) {
		muServer := setupMU(t, func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodPost, r.Method)
			w.WriteHeader(http.StatusInternalServerError)
		})

		ao := NewAOMock("", muServer.URL)
		s := setupSigner(t)

		_, err := ao.SendMessage("testProcess", "testData", nil, "", s)
		assert.Error(t, err)
	})
}

func TestLoadResult_AO(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		cuServer := setupCU(t, func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodGet, r.Method)
			w.WriteHeader(http.StatusOK)
			_, err := w.Write([]byte(`{"Messages": [], "Spawns": [], "Outputs": [], "Error": "", "GasUsed": 0}`))
			assert.NoError(t, err)
		})

		ao := NewAOMock(cuServer.URL, "")
		process := "testProcess"
		message := "testMessage"

		resp, err := ao.LoadResult(process, message)
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, 0, resp.GasUsed)
	})

	t.Run("NonExistentProcessMessage", func(t *testing.T) {
		cuServer := setupCU(t, func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodGet, r.Method)
			w.WriteHeader(http.StatusOK)
			_, err := w.Write([]byte(`{"Messages": [], "Spawns": [], "Outputs": [], "Error": "not found", "GasUsed": 0}`))
			assert.NoError(t, err)
		})

		ao := NewAOMock(cuServer.URL, "")
		_, err := ao.LoadResult("nonExistentProcess", "nonExistentMessage")
		assert.Error(t, err)
	})

	t.Run("HTTPErrorResponse", func(t *testing.T) {
		cuServer := setupCU(t, func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodGet, r.Method)
			w.WriteHeader(http.StatusInternalServerError)
		})

		ao := NewAOMock(cuServer.URL, "")
		_, err := ao.LoadResult("testProcess", "testMessage")
		assert.Error(t, err)
	})
}

func TestDryRun_AO(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		cuServer := setupCU(t, func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodPost, r.Method)
			w.WriteHeader(http.StatusOK)
			_, err := w.Write([]byte(`{"Messages": [], "Spawns": [], "Outputs": [], "Error": "", "GasUsed": 0}`))
			assert.NoError(t, err)
		})

		ao := NewAOMock(cuServer.URL, "")
		message := Message{
			ID:     "testID",
			Target: "testTarget",
			Owner:  "testOwner",
			Data:   "testData",
			Tags:   &[]tag.Tag{},
		}

		resp, err := ao.DryRun(message)
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, 0, resp.GasUsed)
	})

	t.Run("EmptyMessageData", func(t *testing.T) {
		cuServer := setupCU(t, func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodPost, r.Method)
			w.WriteHeader(http.StatusOK)
			_, err := w.Write([]byte(`{"Messages": [], "Spawns": [], "Outputs": [], "Error": "", "GasUsed": 0}`))
			assert.NoError(t, err)
		})

		ao := NewAOMock(cuServer.URL, "")
		message := Message{
			ID:     "testID",
			Target: "testTarget",
			Owner:  "testOwner",
			Data:   "",
			Tags:   &[]tag.Tag{},
		}

		resp, err := ao.DryRun(message)
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, 0, resp.GasUsed)
	})

	t.Run("InvalidMessageFormat", func(t *testing.T) {
		cuServer := setupCU(t, func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodPost, r.Method)
			w.WriteHeader(http.StatusOK)
			_, err := w.Write([]byte(`{"Messages": [], "Spawns": [], "Outputs": [], "Error": "invalid format", "GasUsed": 0}`))
			assert.NoError(t, err)
		})

		ao := NewAOMock(cuServer.URL, "")
		message := Message{
			ID:     "",
			Target: "",
			Owner:  "",
			Data:   "",
			Tags:   &[]tag.Tag{},
		}

		_, err := ao.DryRun(message)
		assert.Error(t, err)
	})

	t.Run("HTTPErrorResponse", func(t *testing.T) {
		cuServer := setupCU(t, func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodPost, r.Method)
			w.WriteHeader(http.StatusInternalServerError)
		})

		ao := NewAOMock(cuServer.URL, "")
		message := Message{
			ID:     "testID",
			Target: "testTarget",
			Owner:  "testOwner",
			Data:   "testData",
			Tags:   &[]tag.Tag{},
		}

		_, err := ao.DryRun(message)
		assert.Error(t, err)
	})
}
