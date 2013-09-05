/* HTML5 Uploader */
/* Taken from http://stackoverflow.com/questions/1219860/javascript-jquery-html-encoding */
function htmlEscape(str) {
    return String(str).replace(/&/g, '&amp;').replace(/"/g, '&quot;').replace(/'/g, '&#39;').replace(/</g, '&lt;').replace(/>/g, '&gt;');
}

function onComplete(dat, f, d, uid)
{
    var dsp = JSON.parse(dat);

    if(Number(dsp.uid) === Number(uid)) {
        if (!dsp.err) {
            var ext = htmlEscape(f.name.split('.').pop().toLowerCase());
            d.innerHTML = "<a href=\"http://s.watfile.com/"+dsp.file+'.'+ext+"\" class=\"uploaded\">http://s.watfile.com/"+dsp.file+'.'+ext+"</a> ( "+htmlEscape(f.name)+" )";
        } else {
            if (dsp.err == "error") {
                d.innerHTML = "<div id=\"error\">Error uploading file! ( "+htmlEscape(f.name)+" )</div>";
            } else if (dsp.err == "size") {
                d.innerHTML = "<div id=\"error\">File is too big! ( "+htmlEscape(f.name)+" )</div>";
            } else if (dsp.err == "rate") {
                d.innerHTML = "<div id=\"error\">Only 6 files per minute! ( "+htmlEscape(f.name)+" )</div>";
            }
        }
        d.innerHTML = "<li>"+d.innerHTML+"</li>";
        document.getElementsByTagName("ul")[0].appendChild(d);
    } else {
        d.innerHTML = "<li><div id=\"error\">Error uploading file! ( "+htmlEscape(f.name)+" )</div></li>";
        document.getElementsByTagName("ul")[0].appendChild(d);
    }
}

function upload(file, xhr) {
    
    var dOutput = document.createElement("ul");
    document.getElementsByTagName("ul")[0].appendChild(dOutput);            
    if(file.size > 10485761)
    {
        dOutput.innerHTML = "<li><div id=\"error\">File is too big! ( "+htmlEscape(file.name)+" )</div></li>";
        return;
    }

    // Firefox 3.6, Chrome 6, WebKit
    if(window.FileReader) { 
        
        dOutput.innerHTML = "<li>Uploading... ( "+htmlEscape(file.name)+" )</li>"; 

        // Once the process of reading file
        this.loadEnd = function() {
            var upid = String(Math.floor(Math.random()*1000000000));
            
            // Firefox 3.6 provides a feature sendAsBinary ()
            if(xhr.sendAsBinary != null) {
                var boundary = 'xxxxxxxxx';
                var body = '--' + boundary + "\r\n";  
                body += "Content-Disposition: form-data; name='upload'; filename='" + file.name + "'\r\n"; 
                body += "Content-Type: application/octet-stream\r\n\r\n";  
                body += reader.result + "\r\n";
                body += '--' + boundary + '--';      
                xhr.open('POST', 'upload', true);
                xhr.setRequestHeader('Content-Type', 'multipart/form-data; boundary=' + boundary);
                xhr.setRequestHeader('X-Requested-With', 'XMLHttpRequest');
                xhr.setRequestHeader('UP-FILENAME', file.name);
                xhr.setRequestHeader('UP-SIZE', file.size);
                xhr.setRequestHeader('UP-TYPE', file.type);
                xhr.setRequestHeader('UP-ID', upid);
                xhr.onreadystatechange = function() { if (xhr.readyState == 4 && xhr.status == 200) { onComplete(xhr.responseText, file, dOutput, upid); }; };
                xhr.upload.addEventListener('progress', loadProgress, false);
                xhr.sendAsBinary(body); 
            // Chrome 7 sends data but you must use the base64_decode on the PHP side
            } else { 
                xhr.open('POST', 'upload?base64=true', true);
                xhr.setRequestHeader('X-Requested-With', 'XMLHttpRequest');
                xhr.setRequestHeader('UP-FILENAME', file.name);
                xhr.setRequestHeader('UP-SIZE', file.size);
                xhr.setRequestHeader('UP-TYPE', file.type);
                xhr.setRequestHeader('UP-ID', upid);
                xhr.onreadystatechange = function() { if (xhr.readyState == 4 && xhr.status == 200) { onComplete(xhr.responseText, file, dOutput, upid); }; };	
                xhr.upload.addEventListener('progress', loadProgress, false);
                xhr.send(window.btoa(reader.result));
            }
        };
            
        // Loading errors
        this.loadError = function(event) {
            switch(event.target.error.code) {
                case event.target.error.NOT_FOUND_ERR:
                    dOutput.innerHTML = '<li>File not found! ( '+htmlEscape(file.name)+' )</li>';
                    break;
                case event.target.error.NOT_READABLE_ERR:
                    dOutput.innerHTML = '<li>File not readable! ( '+htmlEscape(file.name)+' )</li>';
                    break;
                case event.target.error.ABORT_ERR:
                    break; 
                default:
                    dOutput.innerHTML = '<li>Read error! ( '+htmlEscape(file.name)+' )</li>';
            }	
        }
    
        // Reading Progress
        this.loadProgress = function(event) {
        if (event.lengthComputable) {
                var percentage = Math.round((event.loaded * 100) / event.total);
                if (percentage === 100) {
                    dOutput.innerHTML = '<li>Processing... ( '+htmlEscape(file.name)+' )</li>';
                } else  {
                    dOutput.innerHTML = '<li>Uploading : '+percentage+'% ( '+htmlEscape(file.name)+' )</li>';
                }
            }				
        }
            
    var reader = new FileReader();
    // Firefox 3.6, WebKit
    if(reader.addEventListener) {
        reader.addEventListener('loadend', this.loadEnd, false);
        reader.addEventListener('error', this.loadError, false);
    } else {
        reader.onloadend = this.loadEnd;
        reader.onerror = this.loadError;
    }

    // The function that starts reading the file as a binary string
    reader.readAsBinaryString(file);
     
    // Safari 5 does not support FileReader
    } else {
        var upid = Math.floor(Math.random()*1000000000);
        dOutput.innerHTML = "<li>Uploading... ( "+htmlEscape(file.name)+" )</li>"; 
        xhr = new XMLHttpRequest();
        xhr.open('POST', 'upload', true);
        xhr.setRequestHeader('X-Requested-With', 'XMLHttpRequest');
        xhr.setRequestHeader('UP-FILENAME', file.name);
        xhr.setRequestHeader('UP-SIZE', file.size);
        xhr.setRequestHeader('UP-TYPE', file.type);
        xhr.setRequestHeader('UP-ID', upid);
        xhr.onreadystatechange = function() { if (xhr.readyState == 4 && xhr.status == 200) { onComplete(xhr.responseText, file, dOutput, upid); }; };	
        xhr.send(file);
    }				
}

var tid;
function handleFileSelect(evt) {
    evt.stopPropagation();
    evt.preventDefault();
    var files = evt.dataTransfer.files;
    var xhrs = [];
    for (var i = 0; i<files.length; i++) {
        var file = files[i];
        xhrs[i] = new XMLHttpRequest();
        upload(file, xhrs[i]);
    }
    document.getElementById('dropzone').style.display='none';
}
function handleDragOver(evt) {
    evt.stopPropagation();
    evt.preventDefault();
    clearTimeout(tid);
    document.getElementById('dropzone').style.display='block';
}
function handleDragOff(evt) {
    tid=setTimeout(function() {
        evt.stopPropagation();
        document.getElementById('dropzone').style.display='none';
    }, 0);
}

window.addEventListener('dragover', handleDragOver, false);
window.addEventListener('drop', handleFileSelect, false);
window.addEventListener('dragleave', handleDragOff, false);
document.getElementById("elem").onchange = function() { var xhrs = []; for(var i = 0; i<this.files.length; i++) { xhrs[i] = new XMLHttpRequest(); upload(this.files[i], xhrs[i]); } };
document.getElementById("select").addEventListener("click", function (e) { document.getElementById("elem").click(); e.preventDefault(); }, false);
