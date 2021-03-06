package container

import (
	"fmt"
	//"strings"
	"github.com/docker/cli/cli"
	"github.com/docker/cli/cli/command"
	"github.com/docker/docker/api/types"
	//"github.com/docker/cli/opts"
	containertypes "github.com/docker/docker/api/types/container"
	//"github.com/docker/docker/api/types/container"
	//"github.com/pkg/errors"
	"github.com/spf13/cobra"
	//"io/ioutil"
	"context"
	//"go/ast"
	//"github.com/docker/docker/client"
	"log"
	"encoding/json"
	"text/tabwriter"
	"os"
	"time"
	//"github.com/containerd/cgroups"
	"os/exec"
	"strings"
	"strconv"
)
/*
type updateOptions struct {
	blkioWeight        uint16
	cpuPeriod          int64
	cpuQuota           int64
	cpuRealtimePeriod  int64
	cpuRealtimeRuntime int64
	cpusetCpus         string
	cpusetMems         string
	cpuShares          int64
	memory             opts.MemBytes
	memoryReservation  opts.MemBytes
	memorySwap         opts.MemSwapBytes
	kernelMemory       opts.MemBytes
	restartPolicy      string
	cpus               opts.NanoCPUs

	nFlag int

	containers []string
}
*/


type FDSContainer struct{
	ID	string
	containerStats    types.StatsJSON
	//resource	container.Resources
	Period	int64
	Quota	int64
	maxCPUUsage float64
	previousCPU    uint64
	previousSystem uint64
	cpuPercent float64
	historyCPUPer     []float64
	avgCPUPer float64
	needUsage float64
	isNeedCPU bool
}
type allocOptions struct {
	nFlag int
	policy int
}
// NewUpdateCommand creates a new cobra.Command for `docker update`
func NewFDSUpdateCommand(dockerCli command.Cli) *cobra.Command {
//	var options updateOptions
	var options allocOptions







	cmd := &cobra.Command{
		Use:   "FDSAlloc [OPTIONS] ",
		Short: "Turn on the FDS to Dynamiclly allocate CPU resource ",
		Args:  cli.RequiresMinArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			//options.containers = args
			options.nFlag = cmd.Flags().NFlag()
			return runFDSupdate(dockerCli, &options)
		},
	}

	flags := cmd.Flags()
	flags.IntVar(&options.policy, "policy", 0, "FDS allocate policy")
	/*flags.Uint16Var(&options.blkioWeight, "blkio-weight", 0, "Block IO (relative weight), between 10 and 1000, or 0 to disable (default 0)")
	flags.Int64Var(&options.cpuPeriod, "cpu-period", 0, "Limit CPU CFS (Completely Fair Scheduler) period")
	flags.Int64Var(&options.cpuQuota, "cpu-quota", 0, "Limit CPU CFS (Completely Fair Scheduler) quota")
	flags.Int64Var(&options.cpuRealtimePeriod, "cpu-rt-period", 0, "Limit the CPU real-time period in microseconds")
	flags.SetAnnotation("cpu-rt-period", "version", []string{"1.25"})
	flags.Int64Var(&options.cpuRealtimeRuntime, "cpu-rt-runtime", 0, "Limit the CPU real-time runtime in microseconds")
	flags.SetAnnotation("cpu-rt-runtime", "version", []string{"1.25"})
	flags.StringVar(&options.cpusetCpus, "cpuset-cpus", "", "CPUs in which to allow execution (0-3, 0,1)")
	flags.StringVar(&options.cpusetMems, "cpuset-mems", "", "MEMs in which to allow execution (0-3, 0,1)")
	flags.Int64VarP(&options.cpuShares, "cpu-shares", "c", 0, "CPU shares (relative weight)")
	flags.VarP(&options.memory, "memory", "m", "Memory limit")
	flags.Var(&options.memoryReservation, "memory-reservation", "Memory soft limit")
	flags.Var(&options.memorySwap, "memory-swap", "Swap limit equal to memory plus swap: '-1' to enable unlimited swap")
	flags.Var(&options.kernelMemory, "kernel-memory", "Kernel memory limit")
	flags.StringVar(&options.restartPolicy, "restart", "", "Restart policy to apply when a container exits")

	flags.Var(&options.cpus, "cpus", "Number of CPUs")
	flags.SetAnnotation("cpus", "version", []string{"1.29"})
	*/
	return cmd
}

func runFDSupdate(dockerCli command.Cli, options *allocOptions) error {
	client := dockerCli.Client()

	ctx := context.Background()
	var fdsContainer []FDSContainer
	if options.nFlag == 0 {
		options.policy = 1
	}

	FDSOption , err := buildContainerFDSOptions(options)
	if err != nil {
		panic("Error!!")
	}else{
		println(FDSOption.Policy)
	}
	cs, err := dockerCli.Client().ContainerList(ctx, types.ContainerListOptions{})
	if err != nil {
		println("Get Container List Error")
	}else {
		println("Get Container List  OK")
	}
	for _, container := range cs {
		var tmpContainer FDSContainer
		containerInfo, err := client.ContainerInspect(ctx,container.ID)
		if err != nil {
			println("Inspect  Error")
		}
		tmpContainer.ID = containerInfo.ContainerJSONBase.ID
		tmpContainer.Period = containerInfo.ContainerJSONBase.HostConfig.Resources.CPUPeriod
		tmpContainer.Quota = containerInfo.ContainerJSONBase.HostConfig.Resources.CPUQuota
		tmpContainer.maxCPUUsage = float64(tmpContainer.Quota) / float64(tmpContainer.Period)
		tmpContainer.containerStats = initContainerStats(tmpContainer,dockerCli)
		println(tmpContainer.containerStats.PreCPUStats.CPUUsage.TotalUsage)
		fdsContainer = append(fdsContainer,tmpContainer)


	}
	var count=0
	for range time.Tick(time.Millisecond * 1000) {

		for i := 0 ; i < len(cs) ; i++ {
			fdsContainer[i].previousCPU = fdsContainer[i].containerStats.CPUStats.CPUUsage.TotalUsage
			fdsContainer[i].previousSystem = fdsContainer[i].containerStats.CPUStats.SystemUsage
			fdsContainer[i].containerStats = initContainerStats(fdsContainer[i],dockerCli)
			fdsContainer[i].cpuPercent = calculateCPUPercentUnix(fdsContainer[i].previousCPU,fdsContainer[i].previousSystem,&fdsContainer[i].containerStats)
			fdsContainer[i].historyCPUPer = append(fdsContainer[i].historyCPUPer,fdsContainer[i].cpuPercent)
			if len(fdsContainer[i].historyCPUPer) >=5 {
				fdsContainer[i].avgCPUPer = checkAvgCPU(fdsContainer[i].historyCPUPer)
				fdsContainer[i].historyCPUPer = nil
			}
		}
		count+=1
		if count%5 == 0{
			checkUpdateInfo(fdsContainer)
			DoUpdate(dockerCli,fdsContainer)
			count =0
		}
		showInfo(fdsContainer)

	}
	//dockerCli.Client().ContainerStats()
	/*
	containers, err = dockerCli.Client().ContainerFDS(ctx, *FDSOption)
	if err != nil {
		panic(err)
	}else{
		println("func of ContainerFDS is OK!!!")
	}
*/
/*
	for _,cs :=  range  runningContainer {
		print(cs.ID)

	}

*/
	fmt.Println("Add Command is OK!!!")

	return nil
}

func buildContainerFDSOptions(opts *allocOptions) (*types.ContainerFDSOptions, error) {
	options := &types.ContainerFDSOptions{
		Policy:     opts.policy,
	}

	return options, nil
}
func initContainerStats(c FDSContainer, cli command.Cli) types.StatsJSON {
	resp, err := cli.Client().ContainerStats(context.Background(), c.ID, true)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	dec := json.NewDecoder(resp.Body)
	//var status types.StatsJSON
	if err := dec.Decode(&c.containerStats); err != nil {
		log.Fatal(err)
	}
	return c.containerStats
}

func showInfo(containers []FDSContainer) {
	//cleanScreen()
	w := tabwriter.NewWriter(os.Stdout, 12, 1, 3, ' ', 0)
	fmt.Fprint(w, "Name\tCPU%\tAVG\tQuota\tPeriod\tMax_CPU\n")
	for i := 0; i < len(containers); i++ {
		fmt.Fprintf(w, "%s\t%.2f\t%.2f\t%d\t%d\t%.2f\n",
			containers[i].containerStats.Name,
			containers[i].cpuPercent,
			containers[i].avgCPUPer,
			containers[i].Quota,
			containers[i].Period,
			containers[i].maxCPUUsage,

		)

	}

	if err := w.Flush(); err != nil {
		log.Fatal(err)
		return
	}

}

func cleanScreen() {
	cmd := exec.Command("clear")
	cmd.Stdout = os.Stdout
	cmd.Run()

}

func checkAvgCPU(u []float64)float64{
	var total=0.0
	for _, v := range u {
		total+=v
	}
	return (total / float64(5))


}

func checkUpdateInfo (containers []FDSContainer){
	var(
		maxCPUUsage = 1.00
		totalCPUUsage = 0.00
		rtContainerUsage = 0.00
		remainingCPUUsage = 0.00
		rtContainer_num = 0
		needCPUContainer_num = 0


	)

	for _, c := range containers {
		if strings.Contains(c.containerStats.Name, "rt") {
			rtContainerUsage += (c.avgCPUPer / float64(100))
			rtContainer_num++
		} else {
			totalCPUUsage += (c.avgCPUPer / float64(100))
		}

	}
	remainingCPUUsage = maxCPUUsage - 0.5 - totalCPUUsage // rtContainerUsage assumed 0.5

	for i := 0 ; i<len(containers) ; i++{
		if strings.Contains(containers[i].containerStats.Name, "rt") {
			continue
		}
		usage := float64(containers[i].maxCPUUsage - ( containers[i].avgCPUPer / float64(100) ))
		if usage > float64(0.02){
			containers[i].needUsage = (containers[i].avgCPUPer / float64(100)) * float64(1.00) // you can control CPU save
			containers[i].maxCPUUsage = containers[i].needUsage
			containers[i].isNeedCPU = false
			fmt.Println("Noooooooooooooooooooo")
		}else{
			containers[i].isNeedCPU = true
			needCPUContainer_num++
			fmt.Println("NeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeD")

		}

	}

	allocCPU := float64(remainingCPUUsage) / float64(needCPUContainer_num)
	for i := 0 ; i < len(containers) ; i++{
		if containers[i].isNeedCPU{
			containers[i].needUsage = allocCPU + containers[i].maxCPUUsage
			containers[i].maxCPUUsage = containers[i].needUsage

		}
	}

}

func DoUpdate(dockerCli command.Cli,containers []FDSContainer){
	ctx := context.Background()

	for i := 0; i < len(containers); i++ {
		var restartPolicy containertypes.RestartPolicy
		if strings.Contains(containers[i].containerStats.Name, "rt") {
			continue
		}
		allocPeriod := containers[i].Period
		//allocQuota := float64(allocPeriod) * containers[i].needUsage
		allocQuota := containers[i].Quota + 100
		if allocQuota > 1000 {
			resources := containertypes.Resources{
				CPUPeriod:          allocPeriod,
				CPUQuota:           int64(allocQuota),
			}
			updateConfig := containertypes.UpdateConfig{
				Resources:     resources,
				RestartPolicy: restartPolicy,
			}
			_, err := dockerCli.Client().ContainerUpdate(ctx, containers[i].ID, updateConfig)
			if err != nil {
				fmt.Println("Update Error")
			}else{
				fmt.Println("Update OK!!!")
				containers[i].Quota = int64(allocQuota)
				containers[i].isNeedCPU = false

			}



		}




	}





}
func Round(v float64) float64 {
	x, _ := strconv.ParseFloat(fmt.Sprintf("%.3f",v),3)
	return x
}





/*
func runUpdate(dockerCli command.Cli, options *updateOptions) error {
	var err error

	if options.nFlag == 0 {
		return errors.New("you must provide one or more flags when using this command")
	}

	var restartPolicy containertypes.RestartPolicy
	if options.restartPolicy != "" {
		restartPolicy, err = opts.ParseRestartPolicy(options.restartPolicy)
		if err != nil {
			return err
		}
	}

	resources := containertypes.Resources{
		BlkioWeight:        options.blkioWeight,
		CpusetCpus:         options.cpusetCpus,
		CpusetMems:         options.cpusetMems,
		CPUShares:          options.cpuShares,
		Memory:             options.memory.Value(),
		MemoryReservation:  options.memoryReservation.Value(),
		MemorySwap:         options.memorySwap.Value(),
		KernelMemory:       options.kernelMemory.Value(),
		CPUPeriod:          options.cpuPeriod,
		CPUQuota:           options.cpuQuota,
		CPURealtimePeriod:  options.cpuRealtimePeriod,
		CPURealtimeRuntime: options.cpuRealtimeRuntime,
		NanoCPUs:           options.cpus.Value(),
	}

	updateConfig := containertypes.UpdateConfig{
		Resources:     resources,
		RestartPolicy: restartPolicy,
	}

	ctx := context.Background()

	var (
		warns []string
		errs  []string
	)
	for _, container := range options.containers {
		r, err := dockerCli.Client().ContainerUpdate(ctx, container, updateConfig)
		if err != nil {
			errs = append(errs, err.Error())
		} else {
			fmt.Fprintln(dockerCli.Out(), container)
		}
		warns = append(warns, r.Warnings...)
	}
	if len(warns) > 0 {
		fmt.Fprintln(dockerCli.Out(), strings.Join(warns, "\n"))
	}
	if len(errs) > 0 {
		return errors.New(strings.Join(errs, "\n"))
	}
	return nil
}
*/