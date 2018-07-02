package container

import (
	"fmt"
	//"strings"
	"github.com/docker/cli/cli"
	"github.com/docker/cli/cli/command"
	"github.com/docker/docker/api/types"
	//"github.com/docker/cli/opts"
	//containertypes "github.com/docker/docker/api/types/container"
	//"github.com/docker/docker/api/types/container"
	//"github.com/pkg/errors"
	"github.com/spf13/cobra"
	//"io/ioutil"
	"context"
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
		containerInfo, err := client.ContainerInspect(ctx,container.ID)
		if err != nil {
			println("Inspect  Error")
		}
		println(containerInfo.ContainerJSONBase.ID)

	}
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
	fmt.Print("Add Command is OK!!!")

	return nil
}

func buildContainerFDSOptions(opts *allocOptions) (*types.ContainerFDSOptions, error) {
	options := &types.ContainerFDSOptions{
		Policy:     opts.policy,
	}

	return options, nil
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