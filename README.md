# pru
Go library for accessing TI AM335x PRU (Programmable Real-time Unit), which is
available on the BeagleBone Black.

godoc for this package is [available](https://godoc.org/github.com/aamcrae/pru).

This is based on the [beaglebone](https://beagleboard.org/black) [PRU](https://github.com/beagleboard/am335x_pru_package)
package, which contains reference docs for the PRU subsystem, as well as assembler source etc.
If custom PRU programs are to be developed, install the pasm assembler from that package.

[Examples](https://github.com/aamcrae/pru/tree/main/examples) are provided that demonstrate
the API:
 - [swap](https://github.com/aamcrae/pru/tree/main/examples/swap) - a simple program showing how to access
the PRU RAM, and to load and run a simple program.
 - [event](https://github.com/aamcrae/pru/tree/main/examples/event) - a program demonstrating how to use the event processing.
 - [handler](https://github.com/aamcrae/pru/tree/main/examples/handler) - a program showing how to install an asynch event handler.

This is not an officially supported Google product.
