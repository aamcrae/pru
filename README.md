# pru
Go library for accessing TI AM335x PRU (Programmable Real-time Unit), which is
available on the BeagleBone Black.

This library uses UIO to directly access the PRUs, which is only used on older kernels (<=4.19).

Newer kernels use the [RemoteProc](https://software-dl.ti.com/processor-sdk-linux/esd/docs/08_00_00_21/linux/Foundational_Components/PRU-ICSS/Linux_Drivers/RemoteProc.html)
framework for accessing and managing the PRUs.

There is a [newer library](https://github.com/aamcrae/pru-rp)
for Go that uses the RemoteProc framework.

godoc for this package is [available](https://godoc.org/github.com/aamcrae/pru).

This is based on the [beaglebone](https://beagleboard.org/black) [PRU](https://github.com/beagleboard/am335x_pru_package)
package, which contains reference docs for the PRU subsystem, as well as assembler source etc.
If custom PRU programs are to be developed, install the ```pasm``` assembler from that package.

[Examples](https://github.com/aamcrae/pru/tree/main/examples) are provided that demonstrate
the API, including:
 - [swap](https://github.com/aamcrae/pru/tree/main/examples/swap) - a simple program showing how to access
the PRU RAM, and to load and run a simple program.
 - [event](https://github.com/aamcrae/pru/tree/main/examples/event) - a program demonstrating how to use the event processing.
 - [handler](https://github.com/aamcrae/pru/tree/main/examples/handler) - a program showing how to install an asynch event handler.

## Sample skeleton application

```
import "github.com/aamcrae/pru"

func main() {
	// Open and init the PRU subsystem.
	p, _ := pru.Open(pru.DefaultConfig)
	// Get a reference to PRU core #0
	u := p.Unit(0)
	// Get a reference to system event 18
	e := p.Event(18)
	// Run program on PRU 0.
	u.LoadAndRunFile("testprog.bin")
	// Upon completion, the program will send sys event 18 that
	// gets mapped to interrupt device 0, 
	e.Wait()
	p.Close()
}
```

Error handling is omitted for clarity.

## Loading and running PRU programs

PRU programs are normally written in assembler language. An assembler suitable for
compiling PRU programs is installable via the [BeagleBone PRU package](https://github.com/beagleboard/am335x_pru_package).

Once installed, the ```pasm``` utility can used to create PRU programs from PRU assembler source ([Documentation](https://github.com/beagleboard/am335x_pru_package/blob/master/am335xPruReferenceGuide.pdf)).

The PRU is loaded with a binary image containing PRU instruction words.
There is a number of ways of generating and storing these images:
 - A binary image file can be created using the assembler:
```
  pasm -b prucode.p prucode
  # Output binary file is prucode.bin
```
This file can then be loaded and run via the ```LoadAndRunFile``` method:
```
	p := pru.Open()
	u := p.Unit(0)
	u.LoadAndRunFile("prucode.bin")
```
 - The image data can be incorporated as part of the Go program itself by converting the
image data and storing it as a array:
```
   pasm -m prucode.p prucode
   # Output prucode.img
   utils/img2go.sh prucode mypkg
   # prucode_img.go is created with package as mypkg
```
```
	p := pru.Open()
	u := p.Unit(0)
	u.LoadAndRun(prucode_img)
```

These commands can be embedded int the Go source so that the ```go generate``` command
can be used to build the files e.g
```
...
//go:generate pasm -b prucounter.p
...
```
## Accessing Shared Memory

The host CPU can access the various RAM blocks on the PRU subsystem, such as the PRU unit 0 and 1 8KB RAM
and the 12KB shared RAM. These RAM blocks are exported as byte slices (```[]byte```) initialised over the
RAM block as a byte array.

There are a number of ways that applications can access the shared memory as structured access.
For ease of access, the package detects the byte endianess of the PRU subsystem and stores
the order (as a ```binary/encoding Order```) in the PRU structure. This allows use of the ```binary/encoding```
package:

```
	p := pru.Open()
	u := p.Unit(0)
	p.Order.PutUint32(u.Ram[0:], word1)
	p.Order.PutUint32(u.Ram[4:], word2)
	p.Order.PutUint16(u.Ram[offs:], word2)
	...
	v := p.Order.Uint32(u.Ram[20:])
```

Of course, since the RAM is presented as a byte slice, any method that
uses a byte slice can work:

```
	f := os.Open("MyFile")
	f.Read(u.Ram[0x100:0x1FF])
	data := make([]byte, 0x200)
	copy(data, p.SharedRam[0x400:])
```

A Reader/Writer interface is available by using the ```Open``` method on any of the shared RAM fields:

```
	p := pru.Open()
	u := p.Unit(0)
	ram := u.Ram.Open()
	params := []interface{}{
		uint32(event),
		uint32(intrBit),
		uint16(2000),
		uint16(1000),
		uint32(0xDEADBEEF),
		uint32(in),
		uint32(out),
	}
	for _, v := range params {
		binary.Write(ram, p.Order, v)
	}
	...
	ram.Seek(my_offset, io.SeekStart)
	fmt.Fprintf(ram, "Config string %d, %d", c1, c2)
	ram.WriteAt([]byte("A string to be written to PRU RAM"), 0x800)
	ram.Seek(0, io.SeekStart)
	b1 := ram.ReadByte()
	b2 := ram.ReadByte()
	...
```

A caveat is that the RAM is shared with the PRU, and Go does not have any explicit way
of indicating to the compiler that the memory is shared, so potentially there are patterns
of access where the compiler may optimise out accesses if care is not taken - the access may also
be subject to reordering.

If the memory access is done when the PRU units are disabled, then using the Reader/Writer interface or the
```binary/encoding``` methods described above should be sufficient.

For accesses that do rely on explicit ordering and reading or writing, it is recommended that the ```sync/ataomic```
and ```unsafe``` packages are used to access the memory:

```
	p := pru.Open()
	u := p.Unit(0)
	shared_rx := (*uint32)(unsafe.Pointer(&u.Ram[rx_offs]))
	shared_tx := (*uint32)(unsafe.Pointer(&u.Ram[tx_offs]))
	// Load and run PRU program ...
	for {
		n := atomic.LoadUint32(shared_rx)
		// process data from PRU
		...
		// Store word in PRU memory
		atomic.StoreUint32(shared_tx, 0xDEADBEEF)
	}
```

## User-space Event Handling

System events from a range of different sources may be used to trigger
interrupts. There are 64 possible system events, each of which may be enabled or disabled, and
which may be triggered by dedicated hardware, the PRU cores, or the main CPU, depending on the system event.
The system events are mapped to 10 interrupt channels, and these channels may then be mapped to
10 host interrupts.
Whilst 10 host interrupts are available, the first 2 are reserved for sending interrupts to the PRU cores
themselves. The next 8 host interrupts are used to deliver interrupts to the main CPU kernel drivers,
which then make these available via the device interface to the user space applications.

When an event is received, the system event and host interrupts are cleared automatically so that
new events can be received. Multiple system events may be mapped to the same interrupt channel (and thence
to a host interrupt), so when a host interrupt is detected, the set of active system events mapped to that
interrupt channel is retrieved, and a separate event is generated for each active system event.

The [Event](https://pkg.go.dev/github.com/aamcrae/pru#Event)
type is used to access and manage these events via the device interface presented by the kernel drivers.

The two main ways of accessing the signals are:
 - Using the ```Wait``` or ```WaitTimeout``` methods to synchronously
wait upon receiving an event ([example](https://github.com/aamcrae/pru/blob/main/examples/event/event.go))
 - Registering an asynchronous handler that is invoked when a event is received ([example](https://github.com/aamcrae/pru/blob/main/examples/handler/handler.go))

These methods are mutually exclusive - it is not possible to install a handler, and also call ```Wait```
on the same Event.

There are 8 devices ```/dev/uio[0-7]``` that are used to interface user-space to the 8 host interrupts that
are available to the main CPU.

## Configuration

The [PRU Interrupt Controller](https://elinux.org/PRUSSv2_Interrupt_Controller) has a
fairly complex arrangement that allows up to 64 separate system events to be mapped to
up to 10 interrupt channels. These interrupt channels themselves are mapped to
10 host interrupts.

The configuration argument of the ```Open``` function configures the controller
as desired. The configuration contains a mask of PRU units to enable,
mappings of system events to interrupt channels, and
interrupt channels to host interrupts. Mapping a system event will enable that system event
in the interrupt controller, and mapping a channel to a host interrupt will enable that
host interrupt.

At least one PRU core unit must be enabled in the configuration if PRU programs are to be executed.

A default interrupt configuration ```DefaultConfig``` is available.

The default configuration:
 - Enables both PRU core units
 - Assign system events 16 - 25 to interrupt channels 0 - 9
 - Assign interrupt channels 0 - 9 to the corresponding host interrupts 0 - 9

The system events enabled are the events triggered via the PRU Event Interface Mapping
driven via register R31 on the PRU cores.
Host interrupts 0 and 1 are not routed to the ARM CPU, but instead are connected to PRU 0 and 1 respectively.
Host interrupt 2 through 9 are connected to the kernel event devices 0 - 7 respectively (```/dev/uio0``` to ```/dev/uio7```)

## GPIO setup

Considerable documentation is available on the [beaglebone](https://beagleboard.org/) web site
about configuring the GPIO pins. The ```config-pin``` utility can be used to assign selected
GPIO pins to internal PRU registers for easy access:
```
 # config-pin -l P8_11
 Available modes for P8_11 are: default gpio gpio_pu gpio_pd eqep pruout
 # config-pin P8_11 pruout
 Current mode for P8_11 is:     pruout
 # config-pin P8_42 pruin
 Current mode for P8_42 is:     pruin
```
This will assign P8.11 to ```pr1_pru0_pru_r30_15```, which is allocated to bit 15 of PRU unit 0
register 30 (a output GPIO), and P8.42 to ```pr1_pru1_pru_r30_5```, allocated to bit 5 of PRU unit 1
register 31 (an input GPIO).

Using a modified device tree will allow these allocations to be set at boot time.

## Multiple Processes

Multiple Linux processes may access the PRU subsystem concurrently if care is taken. The guidelines are
fairly straighforward:
 - Allocate and enable each PRU core unit to only one process - it may actually be the same process
(indeed this is the default configuration), but it is not possible to share a single PRU core unit
between multiple processes.
 - Do not share interrupt channels or host interrupts between processes; each process must have
a separate host interrupt allocated to that process so that events are delivered reliably to
the process (internally, the events are delivered via reading the ```/dev/uio[0-7]``` device files,
so realistically only 1 process at a time can access each device).

When allowing multiple processes access the PRU concurrently, the configuration used in each process should
reflect how the resources are allocated (i.e PRU core units, host interrupts etc.). For example:
```
  // Config for process 1
  pc := pru.NewConfig()
  // Run code on unit 0, events on host interrupt 4
  pc.EnableUnit(0).Event2Channel(16, 4).Channel2Interrupt(4, 4) 
  ...

  // Config for process 2
  pc := pru.NewConfig()
  // Run code on unit 1
  pc.EnableUnit(1).Event2Channel(17, 3).Channel2Interrupt(3, 3) 
  ...

  // Process 3 does not run any PRU code, but can send and receive system events
  pc := pru.NewConfig()
  pc..Event2Channel(18, 2).Event2Channel(20,2).Channel2Interrupt(2, 2) 
  ...
```

There are no checks to detect if multiple processes are accessing the same resource - the most
likely outcome is the processes treading on each other's control of PRU cores, event
handling and other random behaviour.

## Disclaimer

This is not an officially supported Google product.
