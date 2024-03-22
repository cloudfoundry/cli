package database_test

import (
	"fmt"
	"sync"

	. "code.cloudfoundry.org/cli/integration/assets/hydrabroker/database"
)

var _ = Describe("Database", func() {
	Describe("NewID", func() {
		It("generates IDs of the correct length", func() {
			Expect(NewID()).To(HaveLen(36))
		})

		It("generates a different one every time", func() {
			const times = 1e6
			s := make(map[ID]struct{})
			for i := 0; i < times; i++ {
				s[NewID()] = struct{}{}
			}
			Expect(s).To(HaveLen(times))
		})
	})

	Describe("Database Operations", func() {
		var db *Database

		BeforeEach(func() {
			db = NewDatabase()
		})

		It("can create, retrieve, update and delete", func() {
			By("creating")
			id := NewID()
			data := NewID()
			db.Create(id, data)

			By("retrieving")
			value, ok := db.Retrieve(id)
			Expect(value).To(Equal(data))
			Expect(ok).To(BeTrue())

			By("updating")
			newData := NewID()
			db.Update(id, newData)
			value, ok = db.Retrieve(id)
			Expect(value).To(Equal(newData))
			Expect(ok).To(BeTrue())

			By("deleting")
			db.Delete(id)
			value, ok = db.Retrieve(id)
			Expect(value).To(BeNil())
			Expect(ok).To(BeFalse())
		})

		Describe("Create", func() {
			It("panics on duplicate ID", func() {
				id := NewID()
				db.Create(id, "foo")

				Expect(func() {
					db.Create(id, "foo")
				}).To(PanicWith(fmt.Sprintf("duplicate ID: %s", id)))
			})

		})

		Describe("Update", func() {
			It("panics when the original did not exist", func() {
				id := NewID()
				Expect(func() {
					db.Update(id, "foo")
				}).To(PanicWith(fmt.Sprintf("entry not found: %s", id)))
			})
		})

		Describe("List", func() {
			It("lists current entries", func() {
				id1 := NewID()
				db.Create(id1, id1)
				id2 := NewID()
				db.Create(id2, id2)
				id3 := NewID()
				db.Create(id3, id3)
				db.Delete(id3)

				Expect(db.List()).To(ConsistOf(id1, id2))
			})
		})

		Describe("concurrency", func() {
			It("copes with concurrent create", func() {
				const goroutines = 10
				var wg sync.WaitGroup
				wg.Add(goroutines)

				for i := 0; i < goroutines; i++ {
					go func() {
						defer GinkgoRecover()

						for j := 0; j < 10; j++ {
							Expect(func() {
								id := NewID()
								db.Create(id, "foo")
							}).NotTo(Panic())
						}

						wg.Done()
					}()
				}

				wg.Wait()
			})

			It("copes with concurrent update, retrieve and list", func() {
				const goroutines = 10
				var wg sync.WaitGroup
				id := NewID()
				db.Create(id, "foo")

				wg.Add(goroutines)
				for i := 0; i < goroutines; i++ {
					go func() {
						defer GinkgoRecover()

						for j := 0; j < 10; j++ {
							Expect(func() {
								db.Update(id, "bar")
								value, ok := db.Retrieve(id)
								Expect(value).To(Equal("bar"))
								Expect(ok).To(BeTrue())
								db.List()
							}).NotTo(Panic())
						}

						wg.Done()
					}()
				}

				wg.Wait()
			})

			It("copes with concurrent delete", func() {
				const (
					goroutines = 10
					entries    = 1e4
				)

				ids := make(chan ID, entries)
				for i := 0; i < entries; i++ {
					id := NewID()
					db.Create(id, id)
					ids <- id
				}
				close(ids)

				var wg sync.WaitGroup
				wg.Add(goroutines)
				for i := 0; i < goroutines; i++ {
					go func() {
						defer GinkgoRecover()

						for id := range ids {
							Expect(func() {
								db.Delete(id)
							}).NotTo(Panic())
						}

						wg.Done()
					}()
				}

				wg.Wait()
			})
		})
	})
})
