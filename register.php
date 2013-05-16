<?php
require_once("../includes/watfile/header.php");

if (!(isset($_GET['user']) && isset($_GET['pass']))) {
    echo "missing field";
    exit();
}

$username_us = trim($_GET['user']);
$password_us = trim($_GET['pass']);

if (preg_match('/[a-zA-Z0-9_]{1,15}/i', $username_us) != 1) {
    echo "invalid username";
    exit();
}

$username = $username_us;
$password = password_hash($password_us, PASSWORD_DEFAULT);

if (file_exists($account_dir.$username)) {
    echo "username already exists";
    exit();
}

mkdir($account_dir.$username);
if (write_file_safe($account_dir.$username."/username", $username) === false) { echo "error"; exit(); }
if (write_file_safe($account_dir.$username."/banned", 0) === false) { echo "error"; exit(); }
if (write_file_safe($account_dir.$username."/password", $password) === false) { echo "error"; exit(); }
if (write_file_safe($account_dir.$username."/state", 0) === false) { echo "error"; exit(); }
if (write_file_safe($account_dir.$username."/avatar", "") === false) { echo "error"; exit(); }
if (write_file_safe($account_dir.$username."/views", 0) === false) { echo "error"; exit(); }
if (write_file_safe($account_dir.$username."/created", time()) === false) { echo "error"; exit(); }
write_empty_file($account_dir.$username."/comments");
write_empty_file($account_dir.$username."/list");

login($username);

var_dump($_SESSION);
?>
