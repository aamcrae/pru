# pru
Go library for accessing TI AM335x PRU (Programmable Real-time Unit), which is
available on the BeagleBone Black.

godoc for this package is [available](https://godoc.org/github.com/aamcrae/pru).

This is based on the [beaglebone](https://beagleboard.org/black) [PRU](https://github.com/beagleboard/am335x_pru_package)
package, which contains reference docs for the PRU subsystem, as well as assembler source etc.
If custom PRU programs are to be developed, install the ```pasm``` assembler from that package.

[Examples](https://github.com/aamcrae/pru/tree/main/examples) are provided that demonstrate
the API:
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
	u.RunFile("testprog.bin")
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
This file can then be loaded and run via the ```RunFile``` method:
```
	p := pru.Open()
	u := p.Unit(0)
	u.RunFile("prucode.bin")
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
	u.Run(prucode_img)
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

The configuration argument of the ```Open``` function configures the interrupt controller
as desired. The configuration contains mappings of system events to interrupt channels, and
interrupt channels to host interrupts. Mapping a system event will enable that system event
in the interrupt controller, and mapping a channel to a host interrupt will enable that
host interrupt.

A default interrupt configuration ```DefaultConfig``` is available.

The default configuration:
 - Assign system events 16 - 25 to interrupt channels 0 - 9
 - Assign interrupt channels 0 - 9 to the corresponding host interrupts 0 - 9

The system events enabled are the events triggered via the PRU Event Interface Mapping
driven via register R31 on the PRU cores.
Host interrupts 0 and 1 are not routed to the ARM CPU, but instead are connected to PRU 0 and 1 respectively.
Host interrupt 2 through 9 are connected to the kernel event devices 0 - 7 respectively (```/dev/uio0``` to ```/dev/uio7```)

## Disclaimer

This is not an officially supported Google product.
