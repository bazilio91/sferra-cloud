package handlers_test

import (
	"context"
	"gorm.io/gorm"
	"testing"
	"time"

	"github.com/bazilio91/sferra-cloud/pkg/api/handlers"
	"github.com/bazilio91/sferra-cloud/pkg/api/router"
	"github.com/bazilio91/sferra-cloud/pkg/auth"
	"github.com/bazilio91/sferra-cloud/pkg/testutils"
	"github.com/gin-gonic/gin"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var (
	ctx        context.Context
	testDB     *testutils.TestDBContainer
	jwtManager *auth.JWTManager
	r          *gin.Engine
	DB         *gorm.DB
)

func TestHandlers(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Handlers Suite")
}

var _ = BeforeSuite(func() {
	gin.SetMode(gin.TestMode)
	ctx = context.Background()

	var err error
	testDB, DB, err = testutils.StartTestDB(ctx)
	Expect(err).NotTo(HaveOccurred())

	jwtManager = auth.NewJWTManager(testDB.Config.JWTSecret, time.Hour*24)
	handlers.SetJWTManager(jwtManager)
	handlers.SetConfig(testDB.Config)

	r = router.SetupRouter(jwtManager, testDB.Config)
})

var _ = AfterSuite(func() {
	err := testutils.StopTestDBContainer(ctx, testDB)
	Expect(err).NotTo(HaveOccurred())
})
