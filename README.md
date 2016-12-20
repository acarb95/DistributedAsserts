# DistributedAsserts
A distributed assert library in GoLang. This implements what we call "locally blocking time-based" asserts. These asserts are locally blocking in the sense that they block the thread calling the assert, thus it blocks the local thread. It is time-based because the assert is not run immediately. In order to ensure that the distributed snapshot taken of all the nodes contains data within a reasonable time frame, we utilize physical clocks and schedule the assert to be taken at a specific time. This allows us to reason about when the state was taken from each node and allows for a stronger assertion, i.e. the programmer knows that if the assertion fails, it failed because the state of the system from time t<sub>0</sub> to t<sub>1</sub> was a bad state. 

## Repository Breakdown
The repository is broken down as follows:
- assert: folder contains the library code
- tests: folder containing the testing code

## Run Sample Code
To test the library, go to the test folder and run a test of your choosing. Each test will have a comment of a variable you can change which will trigger the assertion for that test. You can run it once with the assertion and once without the assertion to see the difference in behavior. Each test will have a README that describes in more detail how to run the tests and where the assertion is.

## Note
If running on macOS Sierra and using GoLang 1.6 this will throw a run-time fatal error. This is because GoLang 1.6 is not compatible with macOS Sierra, the programs appear to run fine though. See [here](https://github.com/golang/go/issues/17492).
