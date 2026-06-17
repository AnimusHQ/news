package gates

import "github.com/AnimusHQ/news/internal/shortform"

// StoryboardImageGate enforces §8: storyboard images are imported from
// chatgpt_manual, cover all required scenes, carry valid paths/hashes, are
// approved, passed visual review, and have operator approval recorded.
type StoryboardImageInput struct {
	Manifest       *shortform.StoryboardImageManifest
	RequiredScenes []string
}

func StoryboardImageGate(in StoryboardImageInput) Result {
	e := newEval("storyboard_image")
	if in.Manifest == nil {
		e.require(false, "manifest_missing", "storyboard image manifest is required", "")
		return e.result()
	}
	m := in.Manifest
	e.require(m.Source == "chatgpt_manual", "source_not_chatgpt_manual", "storyboard image source must be chatgpt_manual", "source")
	e.require(present(m.Operator), "operator_approval_missing", "operator must be recorded on storyboard images", "operator")

	byScene := map[string]shortform.StoryboardImage{}
	for _, img := range m.Images {
		byScene[img.SceneID] = img
	}
	for _, scene := range in.RequiredScenes {
		img, ok := byScene[scene]
		if !ok {
			e.require(false, "scene_image_missing", "required scene has no imported image", scene)
			continue
		}
		e.require(present(img.ImagePath), "image_path_missing", "image file path is required", scene)
		e.require(isSHA256(img.ImageHash), "image_hash_missing", "image content hash is required", scene)
		e.require(img.Status == shortform.StatusApproved, "image_not_approved", "storyboard image must be approved", scene)
		e.require(img.VisualReviewPassed, "visual_review_failed", "storyboard image must pass visual review", scene)
	}
	return e.result()
}

// VisualShotGate enforces §8: each shot references an approved storyboard image,
// carries prompt + negative_prompt, provider metadata, output + hash, a duration
// within tolerance, correct scene mapping, and operator approval.
type VisualShotInput struct {
	Manifest            *shortform.VisualShotManifest
	ApprovedImageHashes map[string]bool
	KnownScenes         map[string]bool
	MaxDurationSec      float64
}

func VisualShotGate(in VisualShotInput) Result {
	e := newEval("visual_shot")
	if in.Manifest == nil {
		e.require(false, "manifest_missing", "visual shot manifest is required", "")
		return e.result()
	}
	m := in.Manifest
	e.require(present(m.Provider.Name), "provider_metadata_missing", "visual provider metadata is required", "provider.name")
	maxDuration := in.MaxDurationSec
	if maxDuration <= 0 {
		maxDuration = 60
	}
	for _, shot := range m.Shots {
		field := shot.SceneID
		e.require(present(shot.Prompt), "prompt_missing", "shot prompt is required", field)
		e.require(present(shot.NegativePrompt), "negative_prompt_missing", "shot negative_prompt is required", field)
		e.require(present(shot.OutputPath) && isSHA256(shot.OutputHash), "output_missing", "shot output path and hash are required", field)
		e.require(shot.DurationSec > 0 && shot.DurationSec <= maxDuration, "duration_out_of_tolerance", "shot duration is out of tolerance", field)
		e.require(in.ApprovedImageHashes[shot.ReferenceImageHash], "reference_image_not_approved", "shot must reference an approved storyboard image", field)
		if in.KnownScenes != nil {
			e.require(in.KnownScenes[shot.SceneID], "scene_mapping_invalid", "shot scene_id is not a known scene", field)
		}
		e.require(shot.OperatorApproval, "operator_approval_missing", "shot requires operator approval", field)
	}
	return e.result()
}

// SubtitleGate enforces §8: transcript present, at least one caption file (srt or
// ass), word timestamps, safe-zone and sync checks pass, operator approval.
type SubtitleInput struct {
	Manifest *shortform.SubtitleManifest
}

func SubtitleGate(in SubtitleInput) Result {
	e := newEval("subtitle")
	if in.Manifest == nil {
		e.require(false, "manifest_missing", "subtitle manifest is required", "")
		return e.result()
	}
	m := in.Manifest
	e.require(present(m.TranscriptPath) && isSHA256(m.TranscriptHash), "transcript_missing", "transcript path and hash are required", "transcript")
	e.require(present(m.SRTPath) || present(m.ASSPath), "captions_missing", "at least one caption file (srt or ass) is required", "captions")
	e.require(m.Checks.WordTimestamps, "word_timestamps_missing", "subtitle word timestamps are required", "checks.word_timestamps")
	e.require(m.Checks.SafeZone, "safe_zone_failed", "subtitle safe-zone check must pass", "checks.safe_zone")
	e.require(m.Checks.Sync, "sync_failed", "subtitle sync check must pass", "checks.sync")
	e.require(m.OperatorApproval, "operator_approval_missing", "subtitles require operator approval", "operator_approval")
	return e.result()
}

// RenderGate enforces §8: output + hash present, exact vertical render target,
// audio track present, subtitles burned, and production QA approved.
type RenderInput struct {
	Manifest             *shortform.ShortRenderManifest
	ProductionQADecision string
}

func RenderGate(in RenderInput) Result {
	e := newEval("render")
	if in.Manifest == nil {
		e.require(false, "manifest_missing", "short render manifest is required", "")
		return e.result()
	}
	m := in.Manifest
	e.require(len(m.Outputs) > 0, "outputs_missing", "render must produce at least one output", "outputs")
	for _, out := range m.Outputs {
		field := out.Platform
		e.require(present(out.Path) && isSHA256(out.Hash), "output_hash_missing", "render output path and hash are required", field)
		e.require(out.Resolution == shortform.TargetResolution, "resolution_invalid", "render resolution must be 1080x1920", field)
		e.require(out.Aspect == shortform.TargetAspect, "aspect_invalid", "render aspect must be 9:16", field)
		e.require(out.FPS == shortform.TargetFPS, "fps_invalid", "render fps must match the configured target", field)
		e.require(out.AudioTrack, "audio_track_missing", "render must include an audio track", field)
		e.require(out.SubtitlesBurned, "subtitles_not_readable", "render must include burned-in subtitles", field)
	}
	e.require(in.ProductionQADecision == ProductionQAApproved, "production_qa_not_approved", "production QA must be approved before render is technically valid", "production_qa")
	return e.result()
}

// ReleaseGate enforces §8 / §4.4 / §4.8: a publish manifest with approved
// production QA, human release approval, explicit platforms, intentional
// visibility/schedule, correct AI disclosure, and a successful dry-run.
type ReleaseInput struct {
	PublishManifest      *shortform.UploadPostPublishManifest
	ProductionQADecision string
	DryRunSucceeded      bool
}

func ReleaseGate(in ReleaseInput) Result {
	e := newEval("release")
	if in.PublishManifest == nil {
		e.require(false, "publish_manifest_missing", "publish manifest is required", "")
		return e.result()
	}
	m := in.PublishManifest
	e.require(in.ProductionQADecision == ProductionQAApproved, "production_qa_not_approved", "production QA must be approved before release", "production_qa")
	e.require(m.HumanReleaseApproval, "human_release_approval_missing", "human release approval is required", "human_release_approval")
	e.require(len(m.Platforms) > 0, "platforms_not_explicit", "release platforms must be explicit", "platforms")
	e.require(present(m.Visibility), "visibility_not_set", "release visibility must be set intentionally", "visibility")
	if m.Visibility == "scheduled" {
		e.require(present(m.ScheduledAt), "scheduled_at_missing", "scheduled visibility requires scheduled_at", "scheduled_at")
	}
	// AI disclosure is a blocking release gate (§4.8).
	disclosure := AIDisclosureGate(AIDisclosureInput{
		Required: m.AIDisclosureRequired,
		Text:     m.AIDisclosure,
		Present:  present(m.AIDisclosure),
	})
	if disclosure.Blocked() {
		e.reasons = append(e.reasons, disclosure.Reasons...)
	}
	e.require(in.DryRunSucceeded, "dry_run_failed", "release requires a successful dry-run publish", "dry_run")
	return e.result()
}
