/*
 Navicat Premium Data Transfer

 Source Server         : mongo
 Source Server Type    : MongoDB
 Source Server Version : 70008 (7.0.8)
 Source Host           : localhost:27017
 Source Schema         : douyin

 Target Server Type    : MongoDB
 Target Server Version : 70008 (7.0.8)
 File Encoding         : 65001

 Date: 26/04/2024 20:08:53
*/


// ----------------------------
// Collection structure for message
// ----------------------------
db.getCollection("message").drop();
db.createCollection("message");
db.getCollection("message").createIndex({
    convert_id: NumberInt("1"),
    create_time: NumberInt("1")
}, {
    name: "idx_convertId_createTime"
});
