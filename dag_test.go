package herd_test

import (
	"context"
	"fmt"

	. "github.com/mudler/herd"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("zeroinit dag", func() {
	var g *Graph

	BeforeEach(func() {
		g = DAG()
	})

	Context("simple checks", func() {
		It("orders", func() {
			g.DependOn("A", "B")
			g.DependOn("B", "C")
			g.DependOn("C", "D")
			g.DependOn("D", "E")
			Expect(g.TopoSortedLayers()).To(Equal([][]string{{"E"}, {"D"}, {"C"}, {"B"}, {"A"}}))
		})

		It("orders parallel", func() {
			g.DependOn("A", "B")
			g.DependOn("B", "C")
			g.DependOn("C", "D")
			g.DependOn("D", "E")
			g.DependOn("X", "E")
			Expect(g.TopoSortedLayers()).To(
				Or(
					Equal([][]string{{"E"}, {"D", "X"}, {"C"}, {"B"}, {"A"}}),
					Equal([][]string{{"E"}, {"X", "D"}, {"C"}, {"B"}, {"A"}}),
				),
			)
		})
	})

	Context("Sequential runs", func() {
		It("orders parallel", func() {
			f := ""
			g.Add("foo", WithCallback(func(ctx context.Context) error {
				f += "foo"
				return nil
			}), WithDeps("bar"))
			g.Add("bar", WithCallback(func(ctx context.Context) error {
				f += "bar"
				return nil
			}))
			g.Run(context.Background())
			Expect(f).To(Equal("barfoo"))
		})
	})

	Context("With errors", func() {
		It("fails", func() {
			f := ""

			g.Add("foo", WithCallback(func(ctx context.Context) error {
				return fmt.Errorf("failure")
			}), WithDeps("bar"), FatalOp)

			g.Add("bar",
				WithCallback(func(ctx context.Context) error {
					f += "bar"
					return nil
				}),
			)

			err := g.Run(context.Background())
			Expect(err).To(Equal(fmt.Errorf("failure")))
		})
	})

	Context("Sequential runs, background jobs", func() {
		It("orders parallel", func() {
			testChan := make(chan string)
			f := ""
			g.Add("foo", WithCallback(func(ctx context.Context) error {
				f += "triggered"
				return nil
			}), WithDeps("bar"))
			g.Add("bar", WithCallback(func(ctx context.Context) error {
				<-testChan
				return fmt.Errorf("test")
			}), Background)
			g.Run(context.Background())
			Expect(g.State("bar").Error).ToNot(HaveOccurred())
			Expect(f).To(Equal("triggered"))
			testChan <- "foo"
			Eventually(func() error {
				return g.State("bar").Error
			}).Should(HaveOccurred())
		})
	})

	Context("Weak deps", func() {
		It("runs with weak deps", func() {
			f := ""
			g.Add("foo", WithCallback(func(ctx context.Context) error {
				f += "triggered"
				return nil
			}), WithDeps("bar"), WeakDeps)
			g.Add("bar", WithCallback(func(ctx context.Context) error {
				return fmt.Errorf("test")
			}))

			g.Run(context.Background())
			Expect(f).To(Equal("triggered"))
		})
		It("doesn't run without weak deps", func() {
			f := ""
			foo := ""
			g.Add("foo", WithCallback(func(ctx context.Context) error {
				foo = "triggered"
				return nil
			}), WithDeps("bar"))

			g.Add("fooz", WithCallback(func(ctx context.Context) error {
				f = "nomercy"
				return nil
			}), WithDeps("baz"))

			g.Add("baz", WithCallback(func(ctx context.Context) error {
				return nil
			}))

			g.Add("bar", WithCallback(func(ctx context.Context) error {
				return fmt.Errorf("test")
			}))

			err := g.Run(context.Background())
			Expect(err).ToNot(HaveOccurred())

			Expect(g.State("bar").Error).To(HaveOccurred())
			Expect(f).To(Equal("nomercy"))
			Expect(foo).To(Equal(""))
		})
	})

	Context("init", func() {
		var baz bool
		var foo bool

		BeforeEach(func() {
			baz = false
			foo = false
		})

		It("does not run untied jobs", func() {
			g.Add("baz", WithCallback(func(ctx context.Context) error {
				baz = true
				return nil
			}))

			g.Add("foo", WithCallback(func(ctx context.Context) error {
				foo = true
				return nil
			}))

			err := g.Run(context.Background())
			Expect(err).ToNot(HaveOccurred())
			Expect(foo).To(BeFalse())
			Expect(baz).To(BeFalse())
		})

		It("does run all untied jobs", func() {
			g = DAG(EnableInit)

			g.Add("baz", WithCallback(func(ctx context.Context) error {
				baz = true
				return nil
			}))

			g.Add("foo", WithCallback(func(ctx context.Context) error {
				foo = true
				return nil
			}))

			err := g.Run(context.Background())
			Expect(err).ToNot(HaveOccurred())
			Expect(foo).To(BeTrue())
			Expect(baz).To(BeTrue())
		})
	})
})
