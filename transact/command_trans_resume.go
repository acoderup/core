package transact

import (
	"github.com/acoderup/core/basic"
)

type transactResumeCommand struct {
	tnode *TransNode
}

func (trc *transactResumeCommand) Done(o *basic.Object) error {
	defer o.ProcessSeqnum()
	trc.tnode.checkExeOver()
	return nil
}

func SendTranscatResume(tnode *TransNode) bool {
	return tnode.ownerObj.SendCommand(&transactResumeCommand{tnode: tnode}, true)
}
