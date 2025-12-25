# Speech Service - IPA Comparison

A modular speech processing service that compares pronunciation between audio and text by converting both to IPA (International Phonetic Alphabet) representation.

## Features

### Core Functionality
- **Audio → IPA**: Converts audio to IPA using Whisper fine-tuned model (`neurlang/ipa-whisper-small`)
- **Text → IPA**: Converts text to IPA using gruut phonemizer (same tool used for Whisper training)
- **Phoneme Alignment**: Aligns audio and text phonemes to identify exact differences
- **Stateless Service**: Pure analysis service - no storage, ready to integrate with your main API
- **Detailed Analysis**: Returns phoneme-level data (matches, substitutions, deletions, insertions)
- **Modular Architecture**: Clean separation of concerns with independent components
- **WebM Support**: Handles WebM audio files with automatic conversion
- **CLI Interface**: Simple command-line tool for testing

### Accuracy Improvements (New!)
- **Beam Search**: Uses beam search (5 beams by default) instead of greedy decoding for higher accuracy
- **Language-Specific Hints**: Provides language hints to the model for better recognition
- **Confidence Scores**: Tracks confidence for each transcription to identify low-quality results
- **Audio Quality Detection**: Analyzes SNR, clipping, and silence to warn about quality issues
- **IPA Post-Processing**: Fixes common Whisper mistakes with language-specific correction rules
- **VAD (Voice Activity Detection)**: Automatically trims silence from audio for better results
- **Audio Normalization**: Normalizes volume levels for consistent processing
- **Noise Reduction**: Optional noise reduction for cleaner audio (requires `noisereduce` package)
- **Pre-Emphasis Filter**: Optional high-frequency boost for better consonant recognition
- **Long Audio Chunking**: Automatically processes long audio (>30s) in chunks for better accuracy
- **Ensemble Models**: Support for running multiple models and voting for best results

## Architecture

```
speech-service/
├── src/
│   ├── audio/
│   │   └── loader.py                # Audio loading, conversion, preprocessing, quality detection
│   ├── ipa/
│   │   ├── audio_to_ipa.py          # Whisper-based audio→IPA with beam search & chunking
│   │   ├── text_to_ipa.py           # gruut-based text→IPA
│   │   ├── normalizer.py            # IPA normalization utilities
│   │   ├── aligner.py               # Phoneme-level alignment
│   │   ├── confusion_tracker.py     # Utilities for building confusion matrix (your API uses this)
│   │   ├── post_processor.py        # IPA post-processing and error correction
│   │   └── ensemble.py              # Ensemble approach with multiple models
│   └── models/
│       └── whisper_model.py         # Model loading/caching
├── cli.py                           # CLI entry point
└── requirements.txt                 # Python dependencies
```

## Prerequisites

### System Requirements

1. **Python 3.9-3.11** (Python 3.12 not fully tested with gruut)
2. **FFmpeg** (required for WebM audio conversion)

### Install FFmpeg

**Ubuntu/Debian:**
```bash
sudo apt-get update
sudo apt-get install ffmpeg
```

**macOS:**
```bash
brew install ffmpeg
```

**Windows:**
Download from [ffmpeg.org](https://ffmpeg.org/download.html)

## Installation

### Option 1: Using Conda (Recommended)

```bash
# Create conda environment
conda create -n speech-service python=3.11
conda activate speech-service

# Install dependencies
pip install -r requirements.txt

# Install gruut English language data
pip install gruut[en]
```

### Option 2: Using venv

```bash
# Create virtual environment
python3.11 -m venv venv
source venv/bin/activate  # On Windows: venv\Scripts\activate

# Install dependencies
pip install -r requirements.txt

# Install gruut English language data
pip install gruut[en]
```

### First Run - Model Download

The first time you run the service, it will download the Whisper IPA model (~290MB) from HuggingFace. This is cached locally at `~/.cache/huggingface/` for future use.

## Usage

### Basic Usage

```bash
python cli.py --audio sample.webm --text "Hello world"
```

### Specify Language

```bash
python cli.py --audio sample.webm --text "Bonjour" --language fr
```

### Force CPU Usage

```bash
python cli.py --audio sample.webm --text "Hello world" --device cpu
```

### Command Line Options

- `--audio`: Path to audio file (required) - supports WebM, WAV, MP3, etc.
- `--text`: Text transcript of the audio (required)
- `--language`: Language code (default: `en-us`) - must have corresponding gruut language pack installed
- `--device`: Device for Whisper model (`cpu` or `cuda`, default: auto-detect)

## Output

The CLI displays:

1. **Original inputs**: Audio file info and text
2. **IPA transcriptions**: Both audio (Whisper) and text (gruut) IPA outputs
3. **Normalized versions**: Unicode-normalized IPA for comparison
4. **Statistics**: Match status, lengths, number of differences
5. **Differences table**: Character-level differences (if any)
6. **Visual alignment**: Side-by-side comparison with match markers

Example output:
```
Speech Service - IPA Comparison

📁 Loading audio file...
✓ Loaded audio: 16000Hz, 48000 samples

🎤 Converting audio to IPA with Whisper...
Loading Whisper IPA model: neurlang/ipa-whisper-small
Using device: cuda
Model loaded successfully!
✓ Audio IPA transcription complete

📝 Converting text to IPA with gruut...
✓ Text IPA conversion complete

Results:

┌─────────────────────┬──────────────────────────────┐
│ Source              │ IPA Output                   │
├─────────────────────┼──────────────────────────────┤
│ Original Text       │ Hello world                  │
│ Audio (Whisper)     │ hɛˈloʊ wˈɜrld                │
│ Text (gruut)        │ h ɛ l oʊ w ɜ˞ l d            │
└─────────────────────┴──────────────────────────────┘

Statistics:
Exact Match: False
Audio IPA Length: 13 characters
Text IPA Length: 17 characters
Differences: 8 character(s)

✓ Comparison complete!
```

## Supported Audio Formats

- WebM (automatically converted to WAV)
- WAV
- MP3
- FLAC
- Any format supported by librosa/FFmpeg

## Supported Languages

The system supports any language that has:
1. A gruut language pack installed
2. Support in the Whisper IPA model (99+ languages)

### Installing Additional Languages

```bash
# German
pip install gruut[de]

# Spanish
pip install gruut[es]

# French
pip install gruut[fr]
```

Then use with `--language` flag:
```bash
python cli.py --audio german.webm --text "Hallo Welt" --language de
```

## Troubleshooting

### "FFmpeg not found" error

**Solution:** Install FFmpeg (see Prerequisites section above)

### "gruut language not found" error

**Solution:** Install the language pack:
```bash
pip install gruut[en]  # For English
```

### Model download fails

**Solution:** Check internet connection. The model will download automatically on first run. You can also manually download from HuggingFace:
```bash
# Pre-download the model
python -c "from transformers import WhisperProcessor; WhisperProcessor.from_pretrained('neurlang/ipa-whisper-small')"
```

### Poor audio quality results

The service now includes automatic quality detection and preprocessing:
- **VAD (Voice Activity Detection)**: Automatically enabled to trim silence
- **Audio Normalization**: Automatically normalizes volume levels
- **Quality Warnings**: System warns about clipping, low SNR, or excessive silence

For even better results with noisy audio:
```bash
pip install noisereduce
```

Then use noise reduction in your code:
```python
from src.audio.loader import load_audio_file

audio, sr = load_audio_file("audio.webm", reduce_noise=True)
```

### IPA format differences

The system includes a normalizer to handle Unicode variations and format differences between gruut versions. If you encounter format mismatches, this is expected and the normalizer should handle most cases.

## Advanced Usage

### Improving Accuracy with Advanced Features

#### 1. Using Beam Search (Default)

Beam search is enabled by default with 5 beams. To customize:

```python
from src.ipa.audio_to_ipa import WhisperIPAConverter
from src.audio.loader import load_audio_file

# Load audio
audio, sr = load_audio_file("audio.webm")

# Create converter
converter = WhisperIPAConverter()

# Use more beams for higher accuracy (slower)
ipa = converter.audio_to_ipa(audio, sr, num_beams=10)

# Or use greedy search (faster, less accurate)
ipa = converter.audio_to_ipa(audio, sr, num_beams=1)
```

#### 2. Getting Confidence Scores

```python
# Get confidence score with transcription
ipa, confidence = converter.audio_to_ipa(
    audio, sr,
    return_confidence=True
)

print(f"IPA: {ipa}")
print(f"Confidence: {confidence:.2%}")

# Use confidence to filter low-quality results
if confidence < 0.7:
    print("Warning: Low confidence transcription")
```

#### 3. Audio Quality Detection

```python
from src.audio.loader import AudioLoader

loader = AudioLoader()

# Load with quality check
audio, sr, quality = loader.load_audio_with_quality_check("audio.webm")

print(f"Quality Score: {quality['quality_score']:.1f}/100")
print(f"SNR: {quality['snr_db']:.1f} dB")
print(f"Clipping: {quality['clipping_ratio']*100:.2f}%")

# Check for warnings
for warning in quality['warnings']:
    print(f"⚠️  {warning}")
```

#### 4. Advanced Audio Preprocessing

```python
# Apply all preprocessing options
audio, sr = load_audio_file(
    "audio.webm",
    apply_vad=True,           # Trim silence (default: True)
    normalize=True,           # Normalize volume (default: True)
    reduce_noise=True,        # Reduce background noise (requires noisereduce)
    apply_preemphasis=True,   # Boost high frequencies
    vad_top_db=20             # VAD threshold (lower = more aggressive)
)
```

#### 5. Processing Long Audio

```python
# Automatically chunk audio longer than 30 seconds
ipa = converter.audio_to_ipa_chunked(
    audio, sr,
    chunk_duration=30.0,      # Process in 30-second chunks
    language='en',
    num_beams=5
)
```

#### 6. Ensemble Approach (Experimental)

For maximum accuracy, use multiple models with voting:

```python
from src.ipa.ensemble import EnsembleIPAConverter

# Initialize with multiple models
# Note: Only small model is used by default
ensemble = EnsembleIPAConverter(
    model_names=[
        "neurlang/ipa-whisper-small",
        # Add more models as they become available
    ]
)

# Use voting strategy
ipa = ensemble.audio_to_ipa_ensemble(
    audio, sr,
    strategy='vote'  # or 'confidence' or 'first'
)
```

#### 7. Language-Specific Processing

```python
# Provide language hints for better accuracy
ipa = converter.audio_to_ipa(
    audio, sr,
    language='en',  # Language code
    num_beams=5
)

# The post-processor will apply language-specific corrections
# (e.g., English uses /ɹ/ not /r/, Spanish uses /r/ not /ɹ/)
```

#### 8. Custom Post-Processing

```python
from src.ipa.post_processor import IPAPostProcessor

processor = IPAPostProcessor(language='en')

# Post-process raw IPA output
raw_ipa = "hɛloʊ wɜrld"  # Raw from Whisper
corrected_ipa = processor.post_process(raw_ipa)

# Get correction suggestions
suggestions = processor.suggest_corrections(
    audio_ipa="hɛt",     # What was said
    text_ipa="hɛθ"       # What was expected
)
```

### Performance vs Accuracy Trade-offs

| Configuration | Speed | Accuracy | Use Case |
|--------------|-------|----------|----------|
| `num_beams=1` | Fastest | Good | Real-time applications |
| `num_beams=5` (default) | Fast | Better | General use |
| `num_beams=10` | Slower | Best | High-accuracy needs |
| `reduce_noise=True` | Slow | Best for noisy audio | Poor audio quality |
| Ensemble (multiple models) | Slowest | Best overall | Maximum accuracy needed |

## Development

### Project Structure

- `src/audio/loader.py`: Audio loading, conversion, preprocessing, and quality detection
- `src/ipa/audio_to_ipa.py`: Whisper model integration with beam search and chunking
- `src/ipa/text_to_ipa.py`: gruut integration for text→IPA
- `src/ipa/normalizer.py`: IPA normalization and comparison utilities
- `src/ipa/post_processor.py`: Post-processing and error correction
- `src/ipa/ensemble.py`: Ensemble approach with multiple models
- `src/ipa/aligner.py`: Phoneme-level alignment
- `src/ipa/confusion_tracker.py`: Confusion matrix tracking
- `cli.py`: Command-line interface

### Running Tests

```bash
# TODO: Add tests
pytest tests/
```

## Technical Details

### Audio Processing Pipeline

1. Load audio file (WebM → WAV conversion if needed)
2. Resample to 16kHz mono (Whisper requirement)
3. Process through Whisper IPA model
4. Return IPA transcription

### Text Processing Pipeline

1. Parse text with gruut
2. Extract phonemes for each word
3. Join into IPA string
4. Return IPA transcription

### IPA Normalization

The normalizer handles:
- Unicode normalization (NFC form)
- Character substitutions (e.g., ɡ vs g)
- Whitespace normalization
- Stress marker consistency

## Integration with Your Main API

This service is designed as a **stateless microservice**. Your main API should:

1. **Call this service** to get phoneme analysis data
2. **Store results** in your own database
3. **Track user progress** over time
4. **Build confusion matrices** for personalized feedback

### Using the Confusion Tracker in Your API

The `PhonemeConfusionTracker` class is provided as a utility for your main API:

```python
from src.ipa.confusion_tracker import PhonemeConfusionTracker
from src.ipa.aligner import PhonemeAligner

# In your main API:
# 1. Get alignment from this service
aligner = PhonemeAligner()
alignment = aligner.align(audio_ipa, text_ipa)

# 2. Track it in your database
tracker = PhonemeConfusionTracker(storage_path=f"user_data/{user_id}/confusion.json")
tracker.record_alignment(alignment)
tracker.save()

# 3. Get user statistics
stats = tracker.get_phoneme_stats('p')
# Returns: {'total': 100, 'correct': 50, 'accuracy': 50.0, 'substitutions': {'b': 25, 'g': 25}}
```

### Data Structure

The alignment data returned by `PhonemeAligner` is:
```python
[
    ('p', 'p', 'match'),          # Expected /p/, said /p/ ✓
    ('θ', 't', 'substitute'),      # Expected /θ/, said /t/ ✗
    ('ɹ', '-', 'delete'),          # Expected /ɹ/, didn't say it
    ('-', 'ə', 'insert'),          # Didn't expect anything, said /ə/
]
```

Your main API can use this to build user pronunciation profiles.

## Accuracy Improvement Summary

This version includes **11 major accuracy improvements** over the baseline:

| Improvement | Impact | Cost |
|------------|--------|------|
| Beam search (5 beams) | ⭐⭐⭐ High | Low |
| VAD + Normalization | ⭐⭐⭐ High | None |
| Language hints | ⭐⭐ Medium | None |
| Post-processing | ⭐⭐ Medium | None |
| Audio quality detection | ⭐ Low (diagnostic) | None |
| Confidence scores | ⭐ Low (diagnostic) | None |
| Noise reduction | ⭐⭐⭐ High (noisy audio) | Medium |
| Pre-emphasis | ⭐ Low | None |
| Long audio chunking | ⭐⭐ Medium (long files) | None |
| Ensemble models | ⭐⭐⭐ High | Very High |

**Recommended default**: Beam search + VAD + Normalization + Language hints + Post-processing (enabled by default)

**For noisy audio**: Add `reduce_noise=True` when loading audio

**For maximum accuracy**: Use ensemble approach (warning: slow)

## Context: Integration with Main API

This service is part of a larger system:
- **Main Web API**: Takes ChatGPT responses → converts to audio via ElevenLabs
- **Future feature**: MFA integration for karaoke-style playback (word highlighting with timing)
- **This service**: Pronunciation comparison (audio IPA vs expected IPA) - NO timing needed

MFA (Montreal Forced Aligner) will be used in the main API later for forced alignment with timing, not in this pronunciation comparison service.

## Known Limitations

1. **gruut version mismatch**: Model was trained on gruut 0.6.2 (non-existent version), current is 2.4.0. The normalizer and post-processor mitigate format differences.

2. **No timing information**: This service only compares pronunciation, not timing. Use MFA for forced alignment with timestamps.

3. **Model accuracy**: Even with improvements, Whisper IPA model accuracy depends on:
   - Audio quality (use quality detection to identify issues)
   - Language/accent (provide language hints for better results)
   - Background noise (use noise reduction for noisy audio)

4. **WebM conversion overhead**: WebM files are converted to WAV first, adding processing time.

5. **Noise reduction performance**: The optional noise reduction feature (noisereduce) can be slow for large files. Use only when necessary.

6. **Beam search speed**: Higher beam counts (>5) significantly increase processing time. Balance accuracy vs speed for your use case.

## License

[Add your license here]

## Credits

- Whisper IPA model: [neurlang/ipa-whisper-small](https://huggingface.co/neurlang/ipa-whisper-small)
- gruut phonemizer: [rhasspy/gruut](https://github.com/rhasspy/gruut)
- OpenAI Whisper: [openai/whisper](https://github.com/openai/whisper)
