
## Streams and Streamers

A `Stream` instance manages `Streamer`s and pipes data from one Streamer 
to the next. When started, the Stream will treat each Streamer, one at a 
time and from the first to the last, as a Generator.

### End of File (EOF)

When a Streamer returns `&FileInfo{}, nil`, this signals the end of the
data for that file. A Streamer can do this at any time. EOF will 
propagate all the way down the Stream, until stopped by an EOS call *(see 
below for more info on EOS)*.

### End of Stream (EOS)

When a Streamer returns `nil` in place of a `FileInfo`, this signals the 
End of Stream.

If the Streamer signals EOS while Generating *(see below)*, the Stream 
never calls the Streamer again, and moves onto the next Streamer.

A streamer signaling EOS while Receiving will stop the given file *(if 
any)* from continuing down the Pipe(Streamer) chain. See below for 
further details.

### Generating Streamers

A generator is the topmost active Streamer in the Stream. Once started, 
it does not receive any data and can choose to generate any desired files 
& data, or none.

As an example, the first Streamer will be called with `nil,nil` values 
for `FileInfo` and `[]byte`. This signals to the Streamer to start 
generating it's own files and chunks. Generating is of course *optional*.

When the Streamer returns `nil,nil,nil`, this indicates to the Stream 
that the Streamer has no more files to generate. The next Streamer is 
then called with `nil,nil` and the Stream repeats the process, passing 
along any fileinfo and data generated to all the streamers after the 
Generating Streamer.

### Receiving Streamers

A receiving Streamer is any Streamer *after* the current Generator. They 
receive data from the Streamer before them.

Because we want a Streamer to be able to reduce files *(such as 
concatenating)*, or multiply files *(splitting, exploding, etc)*, 
Streamers can create or reduce files at any time, **with one exception**. 
If a Streamer returns a FileInfo and some data, it must complete that 
file before creating any more files.

This ensures that the implementation for Streamers stays simple.  
Streamers don't *need* to keep track of multiple incoming files and their 
states. They will only ever be receiving **one file** at a time.

If a different file is returned by a Streamer, before it has completed 
the last file it returned *(by signaling EOF)*, an Error has occured and 
the Stream will halt.

Receivers can also choose to signal EOS instead of returning a file and 
data. By doing so, the Stream will not propagate files onto the next 
Streamer. For example, if a JPG Compression Streamer compresses and 
writes the files, then anytime it receives a `*.jpg` file it will return 
EOS. That Streamer will still continue to get all of the data for the JPG 
file, but streamers after it will.

### Receivers becoming Generators

Before returning the given file for the first time, a receiver can opt to 
instead return a new file. By doing this, the Receiver becomes a 
Generator temporarily.

Once a Receiver becomes a Generator, it can generate all the files it 
wishes just as a normal generator. When signaling **EOS**, the Stream 
becomes a Receiver again and the last FileInfo and []byte is 
**repeated**.

This means any Streamer can ignore a file that it hasn't returned *(no 
need to store it locally)*, and Generate it's own files and data and not 
fear losing the original files data. It will repeate, after the Generator 
is signals EOS.

**IMPORTANT**: Becoming a Generator should be reserved for actions like 
concatenation. Where you're buffering multiple files and writing becomes 
important to avoid memory consumption.

If a Streamer simply needs to generate data with no interaction with 
receiving files, then it should return all files and data it receives and 
wait for `nil,nil` to signal it's turn to Generate.

### Errors

