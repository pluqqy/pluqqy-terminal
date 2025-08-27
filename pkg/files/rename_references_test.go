package files

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/pluqqy/pluqqy-cli/pkg/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestRenameComponentUpdatesReferences(t *testing.T) {
	// Create temp directory
	tempDir, err := os.MkdirTemp("", "pluqqy-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)
	
	// Change to temp directory
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(tempDir))
	defer os.Chdir(originalDir)

	// Create directory structure
	componentsDir := filepath.Join(PluqqyDir, ComponentsDir, "contexts")
	pipelinesDir := filepath.Join(PluqqyDir, PipelinesDir)
	archiveComponentsDir := filepath.Join(PluqqyDir, ArchiveDir, ComponentsDir, "contexts")
	archivePipelinesDir := filepath.Join(PluqqyDir, ArchiveDir, PipelinesDir)

	require.NoError(t, os.MkdirAll(componentsDir, 0755))
	require.NoError(t, os.MkdirAll(pipelinesDir, 0755))
	require.NoError(t, os.MkdirAll(archiveComponentsDir, 0755))
	require.NoError(t, os.MkdirAll(archivePipelinesDir, 0755))

	// Create test components
	authComponent := &models.Component{
		Content: "# Auth Context\n\nAuthentication context content",
		Tags:    []string{"auth", "security"},
	}
	userComponent := &models.Component{
		Content: "# User Context\n\nUser context content",
		Tags:    []string{"user"},
	}

	// Write components
	authPath := filepath.Join(ComponentsDir, "contexts", "auth.md")
	userPath := filepath.Join(ComponentsDir, "contexts", "user.md")
	
	require.NoError(t, WriteComponentWithNameAndTags(authPath, authComponent.Content, "Auth Context", authComponent.Tags))
	require.NoError(t, WriteComponentWithNameAndTags(userPath, userComponent.Content, "User Context", userComponent.Tags))

	// Create active pipeline that references both components
	activePipeline := &models.Pipeline{
		Name: "Test Pipeline",
		Path: "test-pipeline.yaml",
		Components: []models.ComponentRef{
			{
				Type:  "contexts",
				Path:  "../components/contexts/auth.md",
				Order: 1,
			},
			{
				Type:  "contexts",
				Path:  "../components/contexts/user.md",
				Order: 2,
			},
		},
		Tags: []string{"test"},
	}

	// Create archived pipeline that references both components
	archivedPipeline := &models.Pipeline{
		Name: "Archived Pipeline",
		Path: "archived-pipeline.yaml",
		Components: []models.ComponentRef{
			{
				Type:  "contexts",
				Path:  "../components/contexts/auth.md",
				Order: 1,
			},
			{
				Type:  "contexts",
				Path:  "../components/contexts/user.md",
				Order: 2,
			},
		},
		Tags: []string{"archived"},
	}

	// Write pipelines
	require.NoError(t, WritePipeline(activePipeline))
	
	// Write archived pipeline manually
	archivedPipelineData, err := yaml.Marshal(archivedPipeline)
	require.NoError(t, err)
	archivedPipelinePath := filepath.Join(PluqqyDir, ArchiveDir, PipelinesDir, "archived-pipeline.yaml")
	require.NoError(t, os.WriteFile(archivedPipelinePath, archivedPipelineData, 0644))

	t.Run("rename active component updates all references", func(t *testing.T) {
		// Rename auth component
		err := RenameComponent(authPath, "Authentication")
		require.NoError(t, err)

		// Verify component was renamed
		newAuthPath := filepath.Join(ComponentsDir, "contexts", "authentication.md")
		_, err = ReadComponent(newAuthPath)
		require.NoError(t, err)

		// Old file should not exist
		_, err = os.Stat(filepath.Join(PluqqyDir, authPath))
		assert.True(t, os.IsNotExist(err))

		// Check active pipeline was updated
		updatedActivePipeline, err := ReadPipeline("test-pipeline.yaml")
		require.NoError(t, err)
		assert.Equal(t, "../components/contexts/authentication.md", updatedActivePipeline.Components[0].Path)
		assert.Equal(t, "../components/contexts/user.md", updatedActivePipeline.Components[1].Path) // Unchanged

		// Check archived pipeline was updated
		updatedArchivedPipeline, err := ReadArchivedPipeline("archived-pipeline.yaml")
		require.NoError(t, err)
		assert.Equal(t, "../components/contexts/authentication.md", updatedArchivedPipeline.Components[0].Path)
		assert.Equal(t, "../components/contexts/user.md", updatedArchivedPipeline.Components[1].Path) // Unchanged
	})

	t.Run("test direct rename of active component referenced in pipelines", func(t *testing.T) {
		// Create a fresh component
		testCompPath := filepath.Join(ComponentsDir, "contexts", "testcomp.md")
		require.NoError(t, WriteComponentWithNameAndTags(testCompPath, "# Test Component\n\nTest content", "Test Component", []string{}))
		
		// Create a pipeline that references it
		testPipeline := &models.Pipeline{
			Name: "Test Direct Pipeline",
			Path: "test-direct.yaml",
			Components: []models.ComponentRef{
				{
					Type:  "contexts",
					Path:  "../components/contexts/testcomp.md",
					Order: 1,
				},
			},
			Tags: []string{},
		}
		require.NoError(t, WritePipeline(testPipeline))
		
		// Rename the component
		err := RenameComponent(testCompPath, "Renamed Test Component")
		require.NoError(t, err)
		
		// Check pipeline was updated
		updatedPipeline, err := ReadPipeline("test-direct.yaml")
		require.NoError(t, err)
		assert.Equal(t, "../components/contexts/renamed-test-component.md", updatedPipeline.Components[0].Path)
	})

	t.Run("rename archived component updates references", func(t *testing.T) {
		// First archive the user component
		userArchivePath := filepath.Join(ArchiveDir, ComponentsDir, "contexts", "user.md")
		userAbsPath := filepath.Join(PluqqyDir, ComponentsDir, "contexts", "user.md")
		userArchiveAbsPath := filepath.Join(PluqqyDir, userArchivePath)
		
		// Archive by moving file
		require.NoError(t, os.Rename(userAbsPath, userArchiveAbsPath))

		// Rename archived component
		err = RenameComponentInArchive(userArchivePath, "User Profile")
		require.NoError(t, err)

		// Verify component was renamed in archive
		newUserArchivePath := filepath.Join(ArchiveDir, ComponentsDir, "contexts", "user-profile.md")
		_, err = ReadComponent(newUserArchivePath)
		require.NoError(t, err)

		// Old archived file should not exist
		_, err = os.Stat(userArchiveAbsPath)
		assert.True(t, os.IsNotExist(err))

		// Check active pipeline - archived components should NOT be automatically updated
		// The user component is now archived but pipelines still point to old path
		// This is the current behavior - we need to fix this
		updatedActivePipeline, err := ReadPipeline("test-pipeline.yaml")
		require.NoError(t, err)
		assert.Equal(t, "../components/contexts/authentication.md", updatedActivePipeline.Components[0].Path) // From previous test
		// TODO: This should be updated but currently isn't - documenting current behavior
		assert.Equal(t, "../components/contexts/user.md", updatedActivePipeline.Components[1].Path)

		// Check archived pipeline
		updatedArchivedPipeline, err := ReadArchivedPipeline("archived-pipeline.yaml")
		require.NoError(t, err)
		assert.Equal(t, "../components/contexts/authentication.md", updatedArchivedPipeline.Components[0].Path) // From previous test
		// TODO: This should be updated but currently isn't - documenting current behavior
		assert.Equal(t, "../components/contexts/user.md", updatedArchivedPipeline.Components[1].Path)
	})

	t.Run("find affected pipelines", func(t *testing.T) {
		// Create a new component
		testComponentPath := filepath.Join(ComponentsDir, "contexts", "test.md")
		require.NoError(t, WriteComponentWithNameAndTags(testComponentPath, "# Test\n\nTest content", "Test", []string{}))

		// Create a pipeline that references it
		pipeline := &models.Pipeline{
			Name: "Test Affected Pipeline",
			Path: "test-affected.yaml",
			Components: []models.ComponentRef{
				{
					Type:  "contexts",
					Path:  "../components/contexts/test.md",
					Order: 1,
				},
			},
			Tags: []string{},
		}
		require.NoError(t, WritePipeline(pipeline))

		// Find affected pipelines
		active, archived, err := FindAffectedPipelines(testComponentPath)
		require.NoError(t, err)
		assert.Contains(t, active, "Test Affected Pipeline")
		assert.Empty(t, archived)
	})

	t.Run("error cases", func(t *testing.T) {
		// Try to rename to existing name
		newTestPath := filepath.Join(ComponentsDir, "contexts", "new-test.md")
		require.NoError(t, WriteComponentWithNameAndTags(newTestPath, "# New Test\n\nContent", "New Test", []string{}))

		err := RenameComponent(newTestPath, "Authentication")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "already exists")

		// Try to rename non-existent component
		err = RenameComponent("components/contexts/nonexistent.md", "Something")
		assert.Error(t, err)
	})
}

func TestRemoveComponentReferencesWithRename(t *testing.T) {
	// Create temp directory
	tempDir, err := os.MkdirTemp("", "pluqqy-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)
	
	// Change to temp directory
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(tempDir))
	defer os.Chdir(originalDir)

	// Create directory structure
	componentsDir := filepath.Join(PluqqyDir, ComponentsDir, "contexts")
	pipelinesDir := filepath.Join(PluqqyDir, PipelinesDir)
	archivePipelinesDir := filepath.Join(PluqqyDir, ArchiveDir, PipelinesDir)

	require.NoError(t, os.MkdirAll(componentsDir, 0755))
	require.NoError(t, os.MkdirAll(pipelinesDir, 0755))
	require.NoError(t, os.MkdirAll(archivePipelinesDir, 0755))

	// Create test components
	comp1Path := filepath.Join(ComponentsDir, "contexts", "comp1.md")
	comp2Path := filepath.Join(ComponentsDir, "contexts", "comp2.md")
	
	require.NoError(t, WriteComponentWithNameAndTags(comp1Path, "# Component 1", "Component 1", []string{}))
	require.NoError(t, WriteComponentWithNameAndTags(comp2Path, "# Component 2", "Component 2", []string{}))

	// Create pipeline with both components
	pipeline := &models.Pipeline{
		Name: "Test Remove Pipeline",
		Path: "test-remove.yaml",
		Components: []models.ComponentRef{
			{
				Type:  "contexts",
				Path:  "../components/contexts/comp1.md",
				Order: 1,
			},
			{
				Type:  "contexts",
				Path:  "../components/contexts/comp2.md",
				Order: 2,
			},
		},
		Tags: []string{},
	}
	require.NoError(t, WritePipeline(pipeline))

	// Remove references to comp1
	err = RemoveComponentReferences(comp1Path)
	require.NoError(t, err)

	// Check pipeline was updated
	updatedPipeline, err := ReadPipeline("test-remove.yaml")
	require.NoError(t, err)
	assert.Len(t, updatedPipeline.Components, 1)
	assert.Equal(t, "../components/contexts/comp2.md", updatedPipeline.Components[0].Path)

	// Remove references to comp2 (pipeline should now be empty)
	err = RemoveComponentReferences(comp2Path)
	require.NoError(t, err)

	// Check pipeline has no components
	updatedPipeline, err = ReadPipeline("test-remove.yaml")
	require.NoError(t, err)
	assert.Len(t, updatedPipeline.Components, 0)
}