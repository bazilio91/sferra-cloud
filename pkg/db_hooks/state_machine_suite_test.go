package db_hooks

import (
	"context"
	"gorm.io/gorm"
	"testing"

	"github.com/bazilio91/sferra-cloud/pkg/testutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var (
	ctx             context.Context
	testDBContainer *testutils.TestDBContainer
	DB              *gorm.DB
)

func TestStateMachine(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "State machine Suite")
}

var _ = BeforeSuite(func() {
	ctx = context.Background()

	var err error
	testDBContainer, DB, err = testutils.StartTestDB(ctx)
	Expect(err).NotTo(HaveOccurred())

	_ = NewStateMachine(DB)
})

var _ = AfterSuite(func() {
	err := testutils.StopTestDBContainer(ctx, testDBContainer)
	Expect(err).NotTo(HaveOccurred())
	DB = nil
})
