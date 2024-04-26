package dbservice

import (
	"math"
	"os"

	"github.com/cloudfoundry/cloud-service-broker/v3/dbservice/models"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var _ = Describe("Migrations", func() {
	DescribeTable(
		"validate last migration",
		func(lastMigration int, expectedError string) {
			err := ValidateLastMigration(lastMigration)
			switch expectedError {
			case "":
				Expect(err).NotTo(HaveOccurred())
			default:
				Expect(err).To(MatchError(expectedError))
			}
		},
		Entry("new-db", -1, nil),
		Entry("before-v2", 0, "migration from broker versions <= 2.0 is no longer supported, upgrade using a v3.x broker then try again"),
		Entry("v3-to-v4", 1, nil),
		Entry("v4-to-v4.1", 2, nil),
		Entry("v4.1-to-v4.2", 3, nil),
		Entry("up-to-date", numMigrations-1, nil),
		Entry("future", numMigrations, "the database you're connected to is newer than this tool supports"),
		Entry("far-future", math.MaxInt32, "the database you're connected to is newer than this tool supports"),
	)

	DescribeTable(
		"RunMigrations() failures",
		func(lastMigration int, expectedError string) {
			db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
			Expect(err).NotTo(HaveOccurred())

			Expect(autoMigrateTables(db, &models.MigrationV1{})).To(Succeed())
			Expect(db.Save(&models.Migration{MigrationID: lastMigration}).Error).To(Succeed())
			Expect(RunMigrations(db)).To(MatchError(expectedError))
		},
		Entry("before-v2", 0, "migration from broker versions <= 2.0 is no longer supported, upgrade using a v3.x broker then try again"),
		Entry("future", numMigrations, "the database you're connected to is newer than this tool supports"),
		Entry("far-future", math.MaxInt32, "the database you're connected to is newer than this tool supports"),
	)

	Describe("RunMigrations() behavior", func() {
		var db *gorm.DB

		BeforeEach(func() {
			var err error
			// The tests don't pass when using an ":memory:" database as opposed to a real file. Presumably a GORM feature.
			db, err = gorm.Open(sqlite.Open("test.sqlite3"), &gorm.Config{})
			Expect(err).NotTo(HaveOccurred())

			DeferCleanup(func() {
				os.Remove("test.sqlite3")
			})
		})

		It("creates a migration table", func() {
			Expect(RunMigrations(db)).To(Succeed())
			Expect(db.Migrator().HasTable("migrations")).To(BeTrue())
		})

		It("applies all migrations when run", func() {
			Expect(RunMigrations(db)).To(Succeed())
			var storedMigrations []models.Migration
			Expect(db.Order("id desc").Find(&storedMigrations).Error).To(Succeed())
			lastMigrationNumber := storedMigrations[0].MigrationID
			Expect(lastMigrationNumber).To(Equal(numMigrations - 1))
		})

		It("can run migrations multiple times", func() {
			for range 10 {
				Expect(RunMigrations(db)).To(Succeed())
			}
		})
	})
})
