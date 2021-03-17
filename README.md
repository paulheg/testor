# Testor

Simple testing tool for commandline based programs

> PS: the code looks really messy because I'm in a hurry, if anyone wants to clean it up, feel free



## Testing

To test a commandline program, simply create a textfile containing the user dialog as following:

````
# this is a comment that will be ignored
# everything starting with > will be sent to the program over std in
# example:
>command-to-pass
# this is the expected response from the program
Hello World
# $ is a regex, it is useful when the exact return value is unknown
# for example:
>generate-random-number
$[0-9]+
````

> test01.txt



To run the the test simply execute the following command on the commandline:

````
testor -testFile="./examples/test01.txt" programToTest.exe argumentForTestProgram
````

If you want to test multiple files you may want to create a script for that.



## Parameters

The program has several parameters:

| Parameter Name | Description                                                  |
| -------------- | ------------------------------------------------------------ |
| testFile       | file path to the testfile                                    |
| logLevel       | level of the logging (default: info)                         |
| cmdPrefix      | prefix for a command that is passed to the program to test (default: >) |
| regexPrefix    | prefix for regex interpretation (default: $)                 |
| commentPrefix  | prefix for comments which are not interpreted (default: #)   |