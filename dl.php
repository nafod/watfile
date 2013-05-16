<?php

require_once("../includes/watfile/header.php");

function absdelete($path)
{
    if (is_dir($path) === true)
    {
        $files = array_diff(scandir($path), array('.', '..'));

        foreach ($files as $file)
        {
            absdelete(realpath($path) . '/' . $file);
        }

        return rmdir($path);
    }

    else if (is_file($path) === true)
    {
        return unlink($path);
    }

    return false;
}

function formatsize($size)
{
    $units = array( 'B', 'KB', 'MB', 'GB', 'TB', 'PB', 'EB', 'ZB', 'YB');
    $power = $size > 0 ? floor(log($size, 1024)) : 0;
    return number_format($size / pow(1024, $power), 2, '.', ',') . ' ' . $units[$power];
}

if (!array_key_exists('id', $_GET))
{
    header("Location: http//watfile.com/");
    exit();
}

if ($_GET['id'] == '/mu-3f8488db-7fabdac2-b1583628-30caf91d') {
    echo '42';
    exit();
}

$argss = explode('/', $_GET['id']);
$args = explode('.', $argss[1]);

if (strlen($args[0]) == 0) {
    header("Location: http://watfile.com/");
    exit();
}

$extra = "";
if (count($argss) > 2) {
    $extra = preg_replace('/[^a-z\d ]/i', '', $argss[2]);
    if (ctype_alnum($extra) === false) {
        header("Location: http://watfile.com/");
        exit();
    }
}

$delete_id = "";
if (count($argss) > 3) {
    $delete_id = preg_replace('/[^a-z\d ]/i', '', $argss[3]);
    if (ctype_alnum($delete_id) === false) {
        header("Location: http://watfile.com/");
        exit();
    }
}

$whitelist = array('image/gif', 'image/png', 'image/jpeg', 'image/bmp', 'application/pdf', 'text/plain');

$id = preg_replace('/[^a-z\d ]/i', '', $args[0]);
if (ctype_alnum($id) === false) {
    header("Location: http://watfile.com/");
    exit();
}

$handle = @opendir($upload_dir.$id.'/');
if($handle == false) {
    header('Location: http://watfile.com/');
    exit();
}

while (false !== ($ifile = readdir($handle))) {
    if ($ifile != '.' && $ifile != '..') {
        break;
    }
}

$file = '/var/www/data-watfile/uploads/'.$id.'/'.$ifile;
$dlonly = 0;
if (file_exists($file)) {
    if(strlen($extra) > 0) {
        if ($extra === "dl") {
            $dlonly = 1;
        } else if ($extra === "info") {
            header('Content-Type: text/plain');
            header('Expires: Sun, 17 Jan 2038 19:14:07 GMT');
            header('Cache-Control: max-age=31536000');
            echo "name: ".base64_decode(basename($file))."\n";
            echo "mime: ".mime_content_type($file)."\n";
            echo "size: ".formatsize(filesize($file))."\n";
            echo "uploaded: ".gmdate('D, d M Y H:i:s T', filemtime($file))."\n";
            echo "md5: ".md5_file($file)."\n";
            echo "sha1: ".sha1_file($file)."\n";
            exit();
        } else if ($extra === "delete" && strlen($delete_id) > 0) {
            if (file_exists('/var/www/data-watfile/delete/'.$delete_id.'/'.$id))
            {
                absdelete('/var/www/data-watfile/delete/'.$delete_id);
                absdelete('/var/www/data-watfile/forcedl/'.$id);
                absdelete('/var/www/data-watfile/hashes/'.md5_file($file));
                absdelete('/var/www/data-watfile/uploads/'.$id);
            }
            header('Location: http://watfile.com/');
            exit();
        }
    }

    header('X-Content-Type-Options: nosniff');
    header('Content-Description: File Transfer');
    $store = mime_content_type($file);
    header('Content-Type: '.mime_content_type($file));
    if(in_array($store, $whitelist) && $dlonly === 0 && !file_exists('/var/www/data-watfile/forcedl/'.$id)) {
        header('Content-Disposition: inline; filename="'.base64_decode(basename($file)).'"');
    } else {
        header('Content-Disposition: attachment; filename="'.base64_decode(basename($file)).'"');
    }
    header('Expires: Sun, 17 Jan 2038 19:14:07 GMT');
    header('Cache-Control: max-age=31536000');
    header('Last-Modified: '.gmdate('D, d M Y H:i:s T', filemtime($file)));
    header('Content-Length: ' . filesize($file));
    header('X-Accel-Redirect: /protected/'.$id.'/'.$ifile);
    ob_clean();
    flush();
}
?>
