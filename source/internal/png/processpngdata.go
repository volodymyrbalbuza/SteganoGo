package pngint

import (
	"encoding/binary"
	"log"
	"os"

	"github.com/Mihai22125/SteganoGo/pkg/png"
)

// extractMetadata extracts data from IHDR chunk. Returns error
func (pngImg *PngImage) extractMetadata(stpng png.StructPNG) error {

	pngImg.meta = imageMetadata{}
	ihdr, err := stpng.IHDRChunk()
	if err != nil {
		//fmt.Println("[extractMetadata]: failed to get IHDR chunk")
		return err
	}

	ihdrData := ihdr.Data()
	err = pngImg.processIHDR(ihdrData)
	if err != nil {
		return err
	}

	return nil
}

// parseIHDR perse IHDR chunk
func (pngImg *PngImage) processIHDR(ihdrData []byte) error {

	pngImg.meta = imageMetadata{}
	if len(ihdrData) != 13 {
		//fmt.Println("IHDR chunk has invalid size")
		return png.ErrInvalidIHDR
	}
	meta := imageMetadata{}

	buf := ihdrData[0:4]
	meta.width = binary.BigEndian.Uint32(buf)
	buf = ihdrData[4:8]
	meta.height = binary.BigEndian.Uint32(buf)
	meta.bitDepth = ihdrData[8]
	meta.colorType = ColorType(ihdrData[9])
	meta.compressionMethod = CompressionMethod(ihdrData[10])
	meta.filterMethod = FiltMethod(ihdrData[11])
	meta.interlaceMethod = InterlaceMethod(ihdrData[12])

	pngImg.meta = meta
	return nil
}

// samplesPerPixel retun bytes per pixel for current image
func (pngImg *PngImage) bytesPerPixel() uint8 {
	bytesPerSample := uint8(0)
	if pngImg.meta.bitDepth < 8 {
		bytesPerSample = 1
	} else {
		bytesPerSample = pngImg.meta.bitDepth / uint8(8)
	}
	return pngImg.samplesPerPixel() * bytesPerSample
}

// samplesPerPixel retun samples per pixel based on color type
func (pngImg *PngImage) samplesPerPixel() uint8 {

	samples := uint8(0)
	switch pngImg.meta.colorType {

	case Grayscale:
		samples = 1
	case IndexedColor:
		samples = 1
	case GrayscaleWithAlpha:
		samples = 2
	case Truecolor:
		samples = 3
	case TruecolorWithAlpha:
		samples = 4
	}

	return samples
}

// stride return bytes per row from png image
func (pngImg *PngImage) stride() uint32 {
	//fmt.Fprintln(os.Stderr, "[stride]: width = ", pngImg.meta.width, " bpp = ", pngImg.bytesPerPixel(), " stride = ", pngImg.meta.width*uint32(pngImg.bytesPerPixel()))
	return pngImg.meta.width * uint32(pngImg.bytesPerPixel())
}

// Unfilter
func (pngImg *PngImage) Unfilter(decompressed []byte) error {

	filterer := newFilterer(pngImg.bytesPerPixel(), uint16(pngImg.stride()), pngImg.meta.height, pngImg.meta.bitDepth)

	// defilter uncompressed data
	err := filterer.reconstruct(decompressed)
	if err != nil {
		return err
	}
	defiltered := filterer.recon
	// assign processed data to png struct
	pngImg.data = defiltered
	return nil
}

// ProcessData consumes an png.StructPNG and it processes png data
func (pngImg *PngImage) ProcessData(stpng *png.StructPNG) error {
	compressor := NewCompressor()

	pngImg.extractMetadata(*stpng)
	IDATdata, err := stpng.IDATdata()
	if err != nil {
		return err
	}
	// decompress png data
	decompressed, err := compressor.DecompressPNGData(IDATdata, pngImg.meta.compressionMethod)
	if err != nil {
		return err
	}

	err = pngImg.Unfilter(decompressed)
	if err != nil {
		return err
	}

	//fmt.Fprintln(os.Stderr, pngImg.data)

	return nil
}

func ProcessImage(path string) (PngImage, error) {
	newPngImage := PngImage{}
	file, err := os.Open(path) // For read access.
	if err != nil {
		log.Println(err)
		return newPngImage, err
	}

	pngData, err := png.ParsePNG(file)
	if err != nil {
		log.Println(err)
		return newPngImage, err
	}
	newPngImage.png = pngData
	newPngImage.ProcessData(&pngData)
	return newPngImage, nil

}
