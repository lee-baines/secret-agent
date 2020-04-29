package memorystore

import (
	"github.com/ForgeRock/secret-agent/pkg/types"
)

// TODO EnsureAcyclic ensures the defined dependencies are acycilic,
//   meaning there are no cirular dependencies

// GetDependencyNodes generates the dependency tree(s)
func GetDependencyNodes(config *types.Configuration) []*types.Node {
	nodes := []*types.Node{}
	// create nodes without parents or children
	nodes = rangeOverSecrets(config.Secrets, nodes, createNode)

	// now set parents and children
	nodes = rangeOverSecrets(config.Secrets, nodes, addParentsAndChildren)

	return nodes
}

// rangeFunc is a function to be run for each path
type rangeFunc func([]string, []string, *types.SecretConfig, *types.KeyConfig, *types.AliasConfig, []*types.Node) []*types.Node

// rangeOverSecrets ranges over the secrets and runs functions to create and update dependency nodes
func rangeOverSecrets(secretsConfig []*types.SecretConfig, nodes []*types.Node, fn rangeFunc) []*types.Node {
	for _, secretConfig := range secretsConfig {
		for _, keyConfig := range secretConfig.Keys {
			// key privateKeyPath
			nodes = fn(keyConfig.PrivateKeyPath, []string{secretConfig.Name, keyConfig.Name}, secretConfig, keyConfig, nil, nodes)
			// key storePassPath
			nodes = fn(keyConfig.StorePassPath, []string{secretConfig.Name, keyConfig.Name}, secretConfig, keyConfig, nil, nodes)
			// key keyPassPath
			nodes = fn(keyConfig.KeyPassPath, []string{secretConfig.Name, keyConfig.Name}, secretConfig, keyConfig, nil, nodes)
			for _, aliasConfig := range keyConfig.AliasConfigs {
				// key alias signedWithPath
				nodes = fn(aliasConfig.SignedWithPath, []string{secretConfig.Name, keyConfig.Name, aliasConfig.Alias}, secretConfig, keyConfig, aliasConfig, nodes)
			}
		}
	}

	return nodes
}

// createNode is a rangeFunc that creates dependency nodes without parents or children
func createNode(parent, path []string, secretConfig *types.SecretConfig, keyConfig *types.KeyConfig, aliasConfig *types.AliasConfig, nodes []*types.Node) []*types.Node {
	// make sure it doesn't already exist
	for _, node := range nodes {
		if Equal(node.Path, path) {
			return nodes
		}
	}
	node := &types.Node{
		Path:         path,
		SecretConfig: secretConfig,
		KeyConfig:    keyConfig,
		AliasConfig:  aliasConfig,
	}
	nodes = append(nodes, node)
	switch len(path) {
	case 2:
		keyConfig.Node = node
	case 3:
		aliasConfig.Node = node
	default:
		panic("Length of path is not 2 or 3!")
	}

	return nodes
}

// addParentsAndChildren is a rangeFunc that sets the parents and children for dependency nodes
func addParentsAndChildren(parent, path []string, secretConfig *types.SecretConfig, keyConfig *types.KeyConfig, aliasConfig *types.AliasConfig, nodes []*types.Node) []*types.Node {
	if len(parent) > 0 {
		// find the parent node(s) of the path
		for _, parentNode := range nodes {
			if Equal(parentNode.Path, parent) {
				// find the node of the path
				for _, node := range nodes {
					if Equal(node.Path, path) {
						node.Parents = append(node.Parents, parentNode)
						parentNode.Children = append(parentNode.Children, node)
					}
				}
			}
		}
	}

	return nodes
}

// Equal checks slice equality
func Equal(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for index, value := range a {
		if value != b[index] {
			return false
		}
	}

	return true
}
