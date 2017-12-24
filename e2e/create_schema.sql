-- MySQL Workbench Forward Engineering

SET @OLD_UNIQUE_CHECKS=@@UNIQUE_CHECKS, UNIQUE_CHECKS=0;
SET @OLD_FOREIGN_KEY_CHECKS=@@FOREIGN_KEY_CHECKS, FOREIGN_KEY_CHECKS=0;
SET @OLD_SQL_MODE=@@SQL_MODE, SQL_MODE='TRADITIONAL,ALLOW_INVALID_DATES';

-- -----------------------------------------------------
-- Schema movie_test_db
-- -----------------------------------------------------
DROP SCHEMA IF EXISTS `movie_test_db` ;

-- -----------------------------------------------------
-- Schema movie_test_db
-- -----------------------------------------------------
CREATE SCHEMA IF NOT EXISTS `movie_test_db` DEFAULT CHARACTER SET utf8 COLLATE utf8_polish_ci ;
USE `movie_test_db` ;

-- -----------------------------------------------------
-- Table `movie_test_db`.`tv_series`
-- -----------------------------------------------------
DROP TABLE IF EXISTS `movie_test_db`.`tv_series` ;

CREATE TABLE IF NOT EXISTS `movie_test_db`.`tv_series` (
  `id` INT UNSIGNED NOT NULL AUTO_INCREMENT,
  `name` VARCHAR(150) NOT NULL,
  `url` VARCHAR(500) NULL,
  PRIMARY KEY (`id`),
  UNIQUE INDEX `name_UNIQUE` (`name` ASC))
ENGINE = InnoDB;


-- -----------------------------------------------------
-- Table `movie_test_db`.`season`
-- -----------------------------------------------------
DROP TABLE IF EXISTS `movie_test_db`.`season` ;

CREATE TABLE IF NOT EXISTS `movie_test_db`.`season` (
  `id` INT UNSIGNED NOT NULL AUTO_INCREMENT,
  `serial_id` INT UNSIGNED NULL,
  `number` INT NOT NULL,
  PRIMARY KEY (`id`),
  INDEX `fk_season_serial_idx` (`serial_id` ASC),
  UNIQUE INDEX `fk_one_season_per_serial_unq` (`serial_id` ASC, `number` ASC),
  CONSTRAINT `fk_season_serial`
    FOREIGN KEY (`serial_id`)
    REFERENCES `movie_test_db`.`tv_series` (`id`)
    ON DELETE NO ACTION
    ON UPDATE NO ACTION)
ENGINE = InnoDB;


-- -----------------------------------------------------
-- Table `movie_test_db`.`episode`
-- -----------------------------------------------------
DROP TABLE IF EXISTS `movie_test_db`.`episode` ;

CREATE TABLE IF NOT EXISTS `movie_test_db`.`episode` (
  `id` INT UNSIGNED NOT NULL AUTO_INCREMENT,
  `season_id` INT UNSIGNED NOT NULL,
  `number` INT NULL,
  `watched` VARCHAR(45) NULL DEFAULT 0,
  `date` DATETIME NULL,
  PRIMARY KEY (`id`),
  INDEX `fk_episode_season_idx` (`season_id` ASC),
  CONSTRAINT `fk_episode_season`
    FOREIGN KEY (`season_id`)
    REFERENCES `movie_test_db`.`season` (`id`)
    ON DELETE NO ACTION
    ON UPDATE NO ACTION)
ENGINE = InnoDB;


SET SQL_MODE=@OLD_SQL_MODE;
SET FOREIGN_KEY_CHECKS=@OLD_FOREIGN_KEY_CHECKS;
SET UNIQUE_CHECKS=@OLD_UNIQUE_CHECKS;

INSERT INTO `tv_series` (`id`,`name`,`url`) VALUES (2,'Arrow','https://www.serialeonline.pl/arrow-online');
INSERT INTO `tv_series` (`id`,`name`,`url`) VALUES (1,'Marvel Agents of Shield','https://www.serialeonline.pl/marvels-agents-of-shield-agenci-tarczy-online');

INSERT INTO `season` (`id`,`serial_id`,`number`) VALUES (1,1,5);
INSERT INTO `season` (`id`,`serial_id`,`number`) VALUES (2,2,6);