package lib

import (
	"fmt"

	"github.com/totegamma/yisp/core"
)

func init() {
	register("k8s", "patch", opPatch)
}

func opPatch(cdr []*core.YispNode, env *core.Env, mode core.EvalMode, e core.Engine) (*core.YispNode, error) {
	if len(cdr) != 2 {
		return nil, core.NewEvaluationError(nil, fmt.Sprintf("patch requires 2 arguments, got %d", len(cdr)))
	}

	targets := cdr[0]
	patches := cdr[1]

	if targets.Kind != core.KindArray || patches.Kind != core.KindArray {
		return nil, core.NewEvaluationError(nil, "patch requires both target and patch to be maps")
	}

	targetArray, ok := targets.Value.([]any)
	if !ok {
		return nil, core.NewEvaluationError(targets, fmt.Sprintf("invalid target type: %T", targets.Value))
	}

	patchArray, ok := patches.Value.([]any)
	if !ok {
		return nil, core.NewEvaluationError(patches, fmt.Sprintf("invalid patch type: %T", patches.Value))
	}

	for _, patchAny := range patchArray {
		patchNode, ok := patchAny.(*core.YispNode)
		if !ok {
			return nil, core.NewEvaluationError(patches, fmt.Sprintf("invalid patch item type: %T", patchAny))
		}

		patchID, err := getManifestID(patchNode)
		if err != nil {
			return nil, core.NewEvaluationErrorWithParent(patchNode, "failed to get GVK from patch", err)
		}

		for i, targetAny := range targetArray {
			targetNode, ok := targetAny.(*core.YispNode)
			if !ok {
				return nil, core.NewEvaluationError(targets, fmt.Sprintf("invalid target item type: %T", targetAny))
			}

			targetID, err := getManifestID(targetNode)
			if err != nil {
				return nil, core.NewEvaluationErrorWithParent(targetNode, "failed to get GVK from target", err)
			}

			if patchID.Covers(targetID) {
				targetArray[i], err = core.DeepMergeYispNode(targetNode, patchNode, targetNode.Type)
				if err != nil {
					return nil, core.NewEvaluationErrorWithParent(patchNode, "failed to apply patch", err)
				}

			}
		}
	}

	return targets, nil
}

type manifestID struct {
	apiVersion string
	kind       string
	namespace  string
	name       string
}

func (id manifestID) String() string {
	if id.namespace != "" {
		return fmt.Sprintf("%s/%s/%s/%s", id.apiVersion, id.kind, id.namespace, id.name)
	}
	return fmt.Sprintf("%s/%s/%s", id.apiVersion, id.kind, id.name)
}

func (id manifestID) Covers(other manifestID) bool {
	if id.apiVersion != "" && id.apiVersion != other.apiVersion {
		return false
	}
	if id.kind != "" && id.kind != other.kind {
		return false
	}
	if id.namespace != "" && id.namespace != other.namespace {
		return false
	}
	if id.name != "" && id.name != other.name {
		return false
	}
	return true
}

func getManifestID(node *core.YispNode) (manifestID, error) {
	var result manifestID
	if node.Kind != core.KindMap {
		return result, fmt.Errorf("expected core.KindMap for GVK, got %s", node.Kind)
	}

	m, ok := node.Value.(*core.YispMap)
	if !ok {
		return result, fmt.Errorf("expected core.YispMap for GVK, got %T", node.Value)
	}

	apiVersionAny, ok := m.Get("apiVersion")
	if ok {
		apiVersionNode, ok := apiVersionAny.(*core.YispNode)
		if !ok {
			return result, fmt.Errorf("expected core.YispNode for apiVersion, got %T", apiVersionAny)
		}
		result.apiVersion, _ = apiVersionNode.Value.(string)
	}

	kindAny, ok := m.Get("kind")
	if ok {
		kindNode, ok := kindAny.(*core.YispNode)
		if !ok {
			return result, fmt.Errorf("expected core.YispNode for kind, got %T", kindAny)
		}
		result.kind, _ = kindNode.Value.(string)
	}

	metadataAny, ok := m.Get("metadata")
	if ok {
		metadataNode, ok := metadataAny.(*core.YispNode)
		if !ok {
			return result, fmt.Errorf("expected core.YispNode for metadata, got %T", metadataAny)
		}
		if metadataNode.Kind != core.KindMap {
			return result, fmt.Errorf("expected core.KindMap for metadata, got %s", metadataNode.Kind)
		}
		metadataMap, ok := metadataNode.Value.(*core.YispMap)
		if !ok {
			return result, fmt.Errorf("expected core.YispMap for metadata, got %T", metadataNode.Value)
		}

		namespaceAny, ok := metadataMap.Get("namespace")
		if ok {
			namespaceNode, ok := namespaceAny.(*core.YispNode)
			if !ok {
				return result, fmt.Errorf("expected core.YispNode for namespace, got %T", namespaceAny)
			}
			result.namespace, _ = namespaceNode.Value.(string)
		}

		nameAny, ok := metadataMap.Get("name")
		if ok {
			nameNode, ok := nameAny.(*core.YispNode)
			if !ok {
				return result, fmt.Errorf("expected core.YispNode for name, got %T", nameAny)
			}
			result.name, _ = nameNode.Value.(string)
		}
	}

	return result, nil
}
