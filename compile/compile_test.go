package main_test

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("Compile", func() {
	var (
		err        error
		binaryPath string
		buildDir   string
	)

	BeforeEach(func() {
		binaryPath, err = Build("github.com/greenhouse-org/hwc-buildpack/compile")
		Expect(err).ToNot(HaveOccurred())

		buildDir, err = ioutil.TempDir("", "")
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		os.RemoveAll(buildDir)
		CleanupBuildArtifacts()
	})

	It("places the web app server in <build_dir>/.cloudfoundry", func() {
		if runtime.GOOS != "windows" {
			Skip("can only be tested on Windows")
		}

		err := ioutil.WriteFile(filepath.Join(buildDir, "Web.config"), []byte("XML"), 0666)
		Expect(err).ToNot(HaveOccurred())

		hwcBinaryPath := filepath.Join(filepath.Dir(binaryPath), "hwc.exe")
		hwcDestPath := filepath.Join(buildDir, ".cloudfoundry", "hwc.exe")
		err = ioutil.WriteFile(hwcBinaryPath, []byte("HWC"), 0666)
		Expect(err).ToNot(HaveOccurred())

		cmd := exec.Command(binaryPath, buildDir, "/cache_dir")
		session, err := Start(cmd, GinkgoWriter, GinkgoWriter)
		Expect(err).ToNot(HaveOccurred())
		Eventually(session).Should(Exit(0))

		_, err = os.Stat(hwcDestPath)
		Expect(err).ToNot(HaveOccurred())

		hwcBinaryContents, err := ioutil.ReadFile(hwcDestPath)
		Expect(err).ToNot(HaveOccurred())
		Expect(string(hwcBinaryContents)).To(Equal("HWC"))
	})

	Context("when not provided any arguments", func() {
		It("fails", func() {
			cmd := exec.Command(binaryPath)
			session, err := Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).ToNot(HaveOccurred())
			Eventually(session).Should(Exit(1))
			Eventually(session.Err).Should(Say("Invalid usage. Expected: compile.exe <build_dir> <cache_dir>"))
		})
	})

	Context("when provided a nonexistent build directory", func() {
		It("fails", func() {
			cmd := exec.Command(binaryPath, "/nonexistent/build_dir", "/cache_dir")
			session, err := Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).ToNot(HaveOccurred())
			Eventually(session).Should(Exit(1))
			Eventually(session.Err).Should(Say("Invalid build directory provided"))
		})
	})

	Context("when the app does not include a Web.config", func() {
		It("fails", func() {
			cmd := exec.Command(binaryPath, buildDir, "/cache_dir")
			session, err := Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).ToNot(HaveOccurred())
			Eventually(session).Should(Exit(1))
			Eventually(session.Err).Should(Say("Missing Web.config"))
		})
	})
})