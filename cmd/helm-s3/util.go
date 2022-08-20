package main

type printer interface {
	Printf(format string, v ...interface{})
	PrintErrf(format string, i ...interface{})
}
