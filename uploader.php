<?php

require_once("../includes/watfile/header.php");

$image_whitelist = array('image/gif', 'image/png', 'image/jpeg', 'image/bmp');

function rate_limit($ip)
{
    global $data_dir;
    if (file_exists($data_dir.'/ratelimit/'.md5($ip)) && filemtime($data_dir.'/ratelimit/'.md5($ip)) + 300 > time())
    {
        $curr = (int)file_get_contents($data_dir.'/ratelimit/'.md5($ip));
        if ($curr === 30)
        {
            return true;
        } else {
            write_file_safe($data_dir.'/ratelimit/'.md5($ip), $curr+1);
        }
    } else {
        write_file_safe($data_dir.'/ratelimit/'.md5($ip), 1);
    }
    return false;
}

function get_hash($hash)
{
    global $data_dir;
    $handle = opendir($data_dir.'/hashes/'.$hash.'/');
    while (false !== ($file = readdir($handle))) {
        if ($file != '.' && $file != '..') {
            break;
        }
    }
    return $file;
}

function make_result($message, $del = '')
{
    if (array_key_exists('HTTP_UP_ID', $_SERVER))
    {
        return $_SERVER['HTTP_UP_ID'].'|'.$del.'|'.$message;
    } else {
        return '0|'.$del.'|'.$message;
    }
}

function unique_id($l = 8, $exists = false) {
    $ret = substr(md5(uniqid(mt_rand(), true)), 0, $l);
    if ($exists === true)
    {
        while(is_dir($upload_dir.$ret)) {
            $ret = substr(md5(uniqid(mt_rand(), true)), 0, $l);
        }
    }
    return $ret;
}

if (!array_key_exists('HTTP_UP_ID', $_SERVER) || !array_key_exists('HTTP_UP_SIZE', $_SERVER) || !array_key_exists('HTTP_UP_FILENAME', $_SERVER))
{
    echo make_result("error");
    exit();
}

if (rate_limit($_SERVER['REMOTE_ADDR']))
{
    echo make_result("rate");
    exit();
}

// If the browser supports sendAsBinary () can use the array $ _FILES
if(count($_FILES)>0) {

    if (!array_key_exists('upload', $_FILES))
    {
        echo make_result("error");
        exit();
    }
  
    if (filesize($_FILES['upload']['tmp_name']) > 10485761)
    {
        echo make_result("size");
        exit();
    } 
    $hash = md5_file($_FILES['upload']['tmp_name']);
    $delete_id = unique_id(30);
    if (is_dir($data_dir.'/hashes/'.$hash))
    {
        $final_id = get_hash($hash);
    } else {
        $final_id = base_convert(unique_id(), 10, 36);
        mkdir($upload_dir.$final_id);
        if (move_uploaded_file( $_FILES['upload']['tmp_name'] , $upload_dir.$final_id.'/'.base64_encode($_FILES['upload']['name']) ) === FALSE) {
            echo make_result("error");
            exit();
        }
        mkdir($data_dir.'/hashes/'.$hash);
        write_empty_file($data_dir.'/hashes/'.$hash.'/'.$final_id);
        mkdir($data_dir.'/delete/'.$delete_id);
        write_empty_file($data_dir.'/delete/'.$delete_id.'/'.$final_id);
        if(in_array(mime_content_type($upload_dir.$final_id.'/'.base64_encode($_FILES['upload']['name'])), $image_whitelist)) {
            if(getimagesize($upload_dir.$final_id.'/'.base64_encode($_FILES['upload']['name'])) === false) {
                write_empty_file($data_dir.'/forcedl/'.$final_id);
            }
        }
    }
    echo make_result($final_id, $delete_id);

} else if(isset($_GET['up'])) {
	// If the browser does not support sendAsBinary ()
	if(isset($_GET['base64'])) {
		$content = base64_decode(file_get_contents('php://input'));
	} else {
		$content = file_get_contents('php://input');
	}
    
    if (mb_strlen($content, '8bit') > 10485761)
    {
        echo make_result("size");
        exit();
    }
 
    $hash = md5($content);
    $delete_id = unique_id(30);
    if (is_dir($data_dir.'/hashes/'.$hash)) {
        $final_id = get_hash($hash);
    } else {
        $final_id = base_convert(unique_id(), 10, 36);
        mkdir($upload_dir.$final_id);
        if (write_file_safe($upload_dir.$final_id.'/'.base64_encode($_SERVER['HTTP_UP_FILENAME']), $content))
        {
            echo make_result("error");
            exit();
        }
        mkdir($data_dir.'/hashes/'.$hash);
        write_empty_file($data_dir.'/hashes/'.$hash.'/'.$final_id);
        mkdir($data_dir.'/delete/'.$delete_id);
        write_empty_file($data_dir.'/delete/'.$delete_id.'/'.$final_id);
        if(in_array(mime_content_type($upload_dir.$final_id.'/'.base64_encode($_SERVER['HTTP_UP_FILENAME'])), $image_whitelist)) {
            if(getimagesize($upload_dir.$final_id.'/'.base64_encode($_SERVER['HTTP_UP_FILENAME'])) === false) {
                write_empty_file($data_dir.'/forcedl/'.$final_id);
            }
        }
    }
    echo make_result($final_id, $delete_id);
}
?>
