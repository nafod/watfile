<?php
require_once("../includes/watfile/header.php");

$username = $_GET['user'];
$password = $_GET['pass'];

if (preg_match('/[a-zA-Z0-9_]{1,15}/i', $username) != 1) {
    header("Location: http://watfile.com/");
    exit();
}

if (is_dir($account_dir.$username))
{
    $password_valid = trim(file_get_contents($account_dir.$username."/password"));
} else {
    echo "no such account!";
    exit();
}

if(password_verify($password, $password_valid) === false) {
    echo "invalid password!";
    exit();
}

login($username);

var_dump($_SESSION);

?>
