package trips

import "strings"

func (se StatusError) Error() string {
	return se.Error()
}

func (se StatusError) Unwrap() error {
	return se.Err
}

func (ge GraphError) Error() string {
	nodeIds := []string(ge)
	var sb strings.Builder
	for i, nodeId := range nodeIds {
		if i > 0 {
			sb.WriteString(",")
		}
		sb.WriteString(nodeId)
	}
	return sb.String()
}

func (mf MultiFirstError) Error() string {
	return GraphError(mf).Error()
}

func (ml MultiLastError) Error() string {
	return GraphError(ml).Error()
}

func (un UnknownNodeIdError) Error() string {
	return GraphError(un).Error()
}

func (ce CycleError) Error() string {
	ges := []GraphError(ce)
	var sb strings.Builder
	for _, ge := range ges {
		sb.WriteString(ge.Error())
		sb.WriteString("\\n")
	}
	return sb.String()
}
