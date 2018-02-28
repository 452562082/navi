//WaveForm struct
typedef struct {
	char riff[4];	// RIFF file identification (4 bytes)
	int length;	// length field (4 bytes)
	char wave[4];	// WAVE chunk identification (4 bytes)
}WAVECHUNK;

typedef struct{
	char fmt[4];	// format sub-chunk identification  (4 bytes)
	int flength;	// length of format sub-chunk (4 byte integer)
	short format;	// format specifier (2 byte integer)
	short chans;	// number of channels (2 byte integer)
	int sampsRate;	// sample rate in Hz (4 byte integer)
	int bpsec;		// bytes per second (4 byte integer)
	short bpsample;	// bytes per sample (2 byte integer)
	short bpchan;	// bits per channel (2 byte integer)
}FMTCHUNK;

typedef struct{
	char szFactID[4];	//'f','a','c','t'
	int dwFactSize;	//the value is 4
}FACTCHUNK;

typedef struct{
	char data[4];	// data sub-chunk identification  (4 bytes)
	int dlength;	// length of data sub-chunk (4 byte integer)
}DATACHUNK;

typedef struct {
	long	length;		// number of samples in the data chunk
	long	samprate;	// sample rate
	long	bitspsample;	// bits per sample
}WAVINFO;

bool WaveSave(char *wavFilename,short *pBuf,long len)
{
	WAVECHUNK	waveChunk;
	FMTCHUNK	fmtChunk;
	DATACHUNK	dataChunk;

	FILE *fp = fopen(wavFilename,"wb");
	if (fp == NULL)
	{
		fclose(fp);
		return false;
	}

	strncpy(waveChunk.riff,"RIFF",4);
	strncpy(waveChunk.wave,"WAVE",4);
	waveChunk.length = len*sizeof(short)/sizeof(char) + 36;
	fwrite(&waveChunk,sizeof(WAVECHUNK),1,fp);

	fmtChunk.bpchan = 16;
	fmtChunk.bpsample = 2;
	fmtChunk.bpsec = 16000;
	fmtChunk.chans = 1;
	fmtChunk.flength = 16;
	fmtChunk.fmt[0]='f',fmtChunk.fmt[1]='m',fmtChunk.fmt[2]='t',fmtChunk.fmt[3]=' ';
	fmtChunk.format = 1;
	fmtChunk.sampsRate = 8000;
	fwrite(&fmtChunk,sizeof(FMTCHUNK),1,fp);

	strncpy(dataChunk.data,"data",4);
	dataChunk.dlength = len*sizeof(short)/sizeof(char);
	fwrite(&dataChunk,sizeof(DATACHUNK),1,fp);

	fwrite(pBuf,sizeof(short),len,fp);

	fclose(fp);

	return true;
}

bool WaveLoad(const char* strFileName,short* &pWavData,unsigned long& len)
{
	int ii,iTotalReaded,iBytesReaded;

	WAVECHUNK	waveChunk;
	FMTCHUNK	fmtChunk;
	DATACHUNK	dataChunk;

	FILE *fp;
	unsigned char *p8;
	short int *p16;
	short *pf;
	char cBuff[0x4000];	// 16K buffer

	if ( (fp=fopen(strFileName,"rb")) == NULL)
	{
		//sprintf(msg,"WavLoad(): Fail to open file \"%s\"",strFileName);
		fclose(fp);
		return false;
	}
	// read the header
	fread(&waveChunk,sizeof(WAVECHUNK),1,fp);

	// check whether it's a wav file
	if ( waveChunk.riff[0] != 'R' || waveChunk.riff[1] != 'I' ||
		waveChunk.riff[2] != 'F' || waveChunk.riff[3] != 'F')
	{
		fclose(fp);
		return false;
	}

	fread(&fmtChunk,sizeof(FMTCHUNK),1,fp);

	if (fmtChunk.chans != 1)
	{
		fclose(fp);
		return false;
	}

	if ( fmtChunk.bpsample != 1 && fmtChunk.bpsample != 2 )
	{
		//sprintf(msg,"WavLoad(): Unresolved sample size --> %d",fmtChunk.bpsample);
		fclose(fp);
		return false;
	}

	fseek(fp,fmtChunk.flength+8-sizeof(FMTCHUNK),SEEK_CUR);

	fread(&dataChunk,sizeof(DATACHUNK),1,fp);

	if (dataChunk.dlength > 0)	// length is valid
	{
		len   = dataChunk.dlength / fmtChunk.bpsample;
		pWavData = new short[len];
	}
	else
	{
		fclose(fp);
		return false;
	}

	if (pWavData==NULL)
	{
		fclose(fp);
		return false;
	}

	pf = pWavData;

	iTotalReaded = 0;
	do {
		iBytesReaded = fread(cBuff,1,0x4000,fp);
		iTotalReaded += iBytesReaded;
		if (iTotalReaded >= dataChunk.dlength)
			iBytesReaded = iBytesReaded - (iTotalReaded-dataChunk.dlength);
		p8  = (unsigned char*)cBuff;
		p16 = (short int *)cBuff;
		for (ii=0;ii<iBytesReaded;ii+=fmtChunk.bpsample)
		{
			if (fmtChunk.bpsample == 1)
				*(pf++) = *(p8++)-128;
			else
				*(pf++) = *(p16++);
		}
	} while (!feof(fp) && (iBytesReaded != 0)
		&& (iTotalReaded < dataChunk.dlength) );

	fclose(fp);
	return true;
}
