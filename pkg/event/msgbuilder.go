package event

import (
	"fmt"
	"regexp"
)

func (k MsgKey) String() string {
	return [...]string{
		"Failed to satisfy policy on resource %s.The following rule(s) %s failed to apply. Created Policy Violation %s",
		"Failed to process rule %s of policy %s. Created Policy Violation %s",
		"Policy %s applied successfully on the resource %s",
		"Rule %s of policy %s applied successful",
		"Failed to apply policy, blocked creation of resource %s. Rule(s) %s failed to apply",
		"Failed to apply rule %s of policy %s. Blocked update of the resource",
		"Failed to apply policy on resource %s. Blocked update of the resource. The following rule(s) %s failed to apply",
	}[k]
}

const argRegex = "%[s,d,v]"

var re = regexp.MustCompile(argRegex)

//GetEventMsg return the application message based on the message id and the arguments,
// if the number of arguments passed to the message are incorrect generate an error
func getEventMsg(key MsgKey, args ...interface{}) (string, error) {
	// Verify the number of arguments
	argsCount := len(re.FindAllString(key.String(), -1))
	if argsCount != len(args) {
		return "", fmt.Errorf("message expects %d arguments, but %d arguments passed", argsCount, len(args))
	}
	return fmt.Sprintf(key.String(), args...), nil
}
