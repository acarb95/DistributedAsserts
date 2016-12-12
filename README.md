# DistributedAsserts
A distributed assert library in GoLang. This implements what we call "locally blocking temporal" asserts. These asserts are locally blocking in the sense that they block the thread calling the assert, thus it blocks the local thread. It is temporal because the assert is not run immediately. In order to ensure that the distributed snapshot taken of all the nodes contains data within a reasonable time frame, we utilize physical clocks and schedule the assert to be taken at a specific time. This allows us to reason about when the state was taken from each node and allows for a stronger assertion, i.e. the programmer knows that if the assertion fails, it failed because the state of the system from time t<sub>0</sub> to t<sub>1</sub> was a bad state. 

## Repository Breakdown
The repository is broken down as follows:
- assert: folder contains the library code
- tests: folder containing the testing code
