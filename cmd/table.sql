/*
SQLyog Community v13.1.6 (64 bit)
MySQL - 5.7.33 : Database - grab
*********************************************************************
*/

/*!40101 SET NAMES utf8 */;

/*!40101 SET SQL_MODE=''*/;

/*!40014 SET @OLD_UNIQUE_CHECKS=@@UNIQUE_CHECKS, UNIQUE_CHECKS=0 */;
/*!40014 SET @OLD_FOREIGN_KEY_CHECKS=@@FOREIGN_KEY_CHECKS, FOREIGN_KEY_CHECKS=0 */;
/*!40101 SET @OLD_SQL_MODE=@@SQL_MODE, SQL_MODE='NO_AUTO_VALUE_ON_ZERO' */;
/*!40111 SET @OLD_SQL_NOTES=@@SQL_NOTES, SQL_NOTES=0 */;
CREATE DATABASE /*!32312 IF NOT EXISTS*/`grab` /*!40100 DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_bin */;

USE `grab`;

/*Table structure for table `order` */

DROP TABLE IF EXISTS `order`;

CREATE TABLE `order` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `order_id` varchar(64) COLLATE utf8mb4_bin NOT NULL COMMENT '订单编号',
  `ticket_id` varchar(64) COLLATE utf8mb4_bin NOT NULL COMMENT '票编号',
  `user_id` varchar(64) COLLATE utf8mb4_bin NOT NULL COMMENT '用户编号',
  `status` tinyint(1) NOT NULL DEFAULT '0' COMMENT '订单状态',
  `deleted` tinyint(1) NOT NULL DEFAULT '0' COMMENT '是否删除',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uniq_order_id` (`order_id`)
) ENGINE=InnoDB AUTO_INCREMENT=59553 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;

/*Table structure for table `ticket_msg` */

DROP TABLE IF EXISTS `ticket_msg`;

CREATE TABLE `ticket_msg` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `ticket_id` varchar(64) COLLATE utf8mb4_bin NOT NULL COMMENT '票编号',
  `ticket_num` bigint(20) NOT NULL COMMENT '票剩余数量',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uniq_ticket_id` (`ticket_id`)
) ENGINE=InnoDB AUTO_INCREMENT=2 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;

/*Table structure for table `ticket_record` */

DROP TABLE IF EXISTS `ticket_record`;

CREATE TABLE `ticket_record` (
  `id` int(5) NOT NULL AUTO_INCREMENT,
  `ticket_id` int(5) NOT NULL,
  `user_id` bigint(10) NOT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;

/*!40101 SET SQL_MODE=@OLD_SQL_MODE */;
/*!40014 SET FOREIGN_KEY_CHECKS=@OLD_FOREIGN_KEY_CHECKS */;
/*!40014 SET UNIQUE_CHECKS=@OLD_UNIQUE_CHECKS */;
/*!40111 SET SQL_NOTES=@OLD_SQL_NOTES */;
