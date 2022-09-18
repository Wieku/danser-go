package hacks

// Intel has some weird problems with DSA glBufferSubData that causes Draw calls not to be flushed immediately,
// resulting in broken image in some cases
var IsIntel bool

// Old AMD drivers on TeraScale GPUs (15.201.xxx.xxx) have some weird problems with texture upload.
// To work correctly, it has to be uploaded on another unit than the one that's frequently used for rendering
var IsOldAMD bool
