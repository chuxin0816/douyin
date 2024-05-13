douyin = db.getSiblingDB('douyin');
douyin.getCollection("message").drop();
douyin.createCollection("message");
douyin.getCollection("message").createIndex({
    convert_id: NumberInt("1"),
    create_time: NumberInt("1")
}, {
    name: "idx_convertId_createTime"
});
