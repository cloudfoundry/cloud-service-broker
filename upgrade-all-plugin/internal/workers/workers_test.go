package workers_test

import (
	"regexp"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/cloudfoundry/cloud-service-broker/upgrade-all-plugin/internal/workers"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Workers", func() {
	It("runs many workers and waits for them to complete", func() {
		var count int32

		workers.Run(3, func() {
			time.Sleep(time.Millisecond)
			atomic.AddInt32(&count, 1)
		})

		Expect(count).To(Equal(int32(3)))
	})

	It("runs the workers in different goroutines", func() {
		const workerCount = 10
		reg := regexp.MustCompile(`^goroutine \d+ `)
		routineIDMap := sync.Map{}

		workers.Run(workerCount, func() {
			buf := make([]byte, 100)
			runtime.Stack(buf, false)
			routineIDMap.Store(string(reg.Find(buf)), true)
		})

		count := 0
		routineIDMap.Range(func(_, _ any) bool {
			count++
			return true
		})

		Expect(count).To(Equal(workerCount))
	})

})
