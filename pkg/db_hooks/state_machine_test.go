package db_hooks

import (
	"context"
	"encoding/json"

	"github.com/bazilio91/sferra-cloud/pkg/proto"
	types2 "github.com/infobloxopen/protoc-gen-gorm/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"gorm.io/datatypes"
)

var _ = Describe("StateMachine", func() {
	var (
		sm *StateMachine
	)

	BeforeEach(func() {
		sm = NewStateMachine(DB)
	})

	Context("Recognition Completed Handler", func() {
		It("should flatten recognition result and update task status", func() {
			// Create a test tree structure
			leaf1 := &proto.TreeNode{
				Id: "leaf1",
			}
			leaf2 := &proto.TreeNode{
				Id: "leaf2",
			}
			root := proto.TreeNode{
				Id:     "root",
				Leaves: []*proto.TreeNode{leaf1, leaf2},
			}

			// Create JSON types
			recognitionResult := datatypes.NewJSONType(root)

			// Create a test task with recognition result
			task, err := createTestTask(DB, proto.Status_STATUS_RECOGNITION_COMPLETED, 100, []string{"test.jpg"}, nil)
			Expect(err).NotTo(HaveOccurred())

			task.RecognitionResult = &recognitionResult
			task.FrontendResult = &datatypes.JSONType[proto.TreeNode]{} // empty but not nil
			task.FrontendResultUnrecognized = &types2.Jsonb{RawMessage: []byte(`{"other": "data"}`)}

			// Save the task
			err = DB.Create(task).Error
			Expect(err).NotTo(HaveOccurred())

			// Call the handler
			err = sm.handleRecognitionCompleted(context.Background(), task)
			Expect(err).NotTo(HaveOccurred())

			// Verify the task state changes
			Expect(task.Status).To(Equal(int32(proto.Status_STATUS_PROCESSING_COMPLETED)))
			Expect(task.FrontendResult).To(BeNil())
			Expect(task.FrontendResultUnrecognized).To(BeNil())

			// Verify the flattened result
			var flattenedNodes []proto.TreeNode
			err = json.Unmarshal(task.FrontendResultFlat.RawMessage, &flattenedNodes)
			Expect(err).NotTo(HaveOccurred())

			// We should have 3 nodes in total (root + 2 leaves)
			Expect(flattenedNodes).To(HaveLen(3))

			// Convert nodes to map for easier assertion
			nodesMap := make(map[string]proto.TreeNode)
			for _, node := range flattenedNodes {
				nodesMap[node.Id] = node
			}

			// Verify parent-child relationships
			Expect(nodesMap).To(HaveKey("root"))
			Expect(nodesMap).To(HaveKey("leaf1"))
			Expect(nodesMap).To(HaveKey("leaf2"))

			leaf1Node := nodesMap["leaf1"]
			leaf2Node := nodesMap["leaf2"]
			Expect(leaf1Node.ParentId).To(Equal("root"))
			Expect(leaf2Node.ParentId).To(Equal("root"))
		})
	})
})
