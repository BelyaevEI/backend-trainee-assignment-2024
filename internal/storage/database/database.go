package database

import (
	"context"
	"database/sql"
	"errors"

	"github.com/BelyaevEI/backend-trainee-assignment-2024/internal/models"
	_ "github.com/jackc/pgx/v5/stdlib"
)

// Implementation check
var _ DBaser = (*dbase)(nil)

type DBaser interface {
	CreateBanner(ctx context.Context, bannerBody models.BannerBody) (int, error)
	UpdateBanner(ctx context.Context, bannerBody models.BannerBody, bannerID int) (bool, error)
	GetBanners(ctx context.Context, queryParam models.Query) ([]models.ResponseBody, error)
	GetBanner(ctx context.Context, featureID, tagID int) (models.BannerContent, error)
	DeleteBanner(ctx context.Context, bannerID int) error
	GetHistoryBanner(ctx context.Context, bannerID int) ([]models.BannerHistory, error)
	UpdateVersion(ctx context.Context, bannerVersion models.BannerHistory) error
}

// Database layer
type dbase struct {
	db *sql.DB
}

func Connect(dsn string) (DBaser, error) {

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	// Create table for actual_banner
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS actual_banner
					(banner_id BIGSERIAL PRIMARY KEY,
					is_active boolean DEFAULT TRUE,
					title text NOT NULL,
					text text NOT NULL,
					url text NOT NULL)`)
	if err != nil {
		return nil, err
	}

	// Create table for history_banner
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS history_banner
					(banner_id bigint NOT NULL,
					version int NOT NULL,
					title text NOT NULL,
					text text NOT NULL,
					url text NOT NULL,
					PRIMARY KEY (banner_id, version))`)
	if err != nil {
		return nil, err
	}

	// Create table for tag and feature
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS tag_feature
					(feature_id bigint NOT NULL,
					tag_id int NOT NULL,
					banner_id bigint NOT NULL,
					PRIMARY KEY (feature_id, tag_id))`)
	if err != nil {
		return nil, err
	}

	return dbase{
		db: db,
	}, nil

}

// 1. Пытаемся сделать запись в actual_banner с title text url
// Если запись удачна, значит переданные данные баннера валидны и получим ID баннера

// 2. Далее нам нужно убедиться, что в таблице tag_feature нет записи с переданными tag и feature
// Если такая запись есть, значит откатываем все изменения, т.к. по условию тэг и фича явно определяют баннер

// 3. И наконец делаем первую запись в history_banner для переданного баннера
func (d dbase) CreateBanner(ctx context.Context, bannerBody models.BannerBody) (int, error) {

	var id int

	tx, err := d.db.Begin()
	if err != nil {
		return 0, err
	}

	defer tx.Rollback()

	// 1. Делаем обновления таблицы actual_banner
	// Не смотрим поле is_active т.к. оно не является телом баннера, а является лишь свойством
	// stmt, err := tx.Prepare("INSERT INTO actual_banner (title, text, url) VALUES ($1, $2, $3)")
	// if err != nil {
	// 	return 0, err
	// }

	// res, err := stmt.ExecContext(ctx, bannerBody.Content.Title, bannerBody.Content.Text, bannerBody.Content.Url)
	// if err != nil {
	// 	return 0, err
	// }

	// // Получаем ID, который инкрементируется в БД при вставке уникального баннера
	// id, err := res.LastInsertId()
	// if err != nil {
	// 	return 0, err
	// }

	err = tx.QueryRowContext(ctx, `INSERT INTO actual_banner (title, text, url) 
									VALUES ($1, $2, $3) RETURNING banner_id`,
		bannerBody.Content.Title,
		bannerBody.Content.Text,
		bannerBody.Content.Url).Scan(&id)
	if err != nil {
		return 0, err
	}

	// 2. Делаем обновления таблицы tag_feature
	_, err = tx.ExecContext(ctx, `INSERT INTO tag_feature 
								(feature_id, tag_id, banner_id) 
								VALUES($1, $2, $3)`,
		bannerBody.FeatureID,
		bannerBody.TagID,
		id,
	)
	if err != nil {
		return 0, err
	}

	// 3. Делаем первую запись в таблицу history_banner
	_, err = tx.ExecContext(ctx, `INSERT INTO history_banner
								(banner_id, version, title, text, url)
								VALUES($1, $2, $3, $4, $5)`,
		id,
		1,
		bannerBody.Content.Title,
		bannerBody.Content.Text,
		bannerBody.Content.Url,
	)
	if err != nil {
		return 0, err
	}

	err = tx.Commit()
	if err != nil {
		return 0, err
	}

	return int(id), nil
}

// 1. Смотрим в таблицу actual_banner - существует ли запись с таким ID(существует ли баннер)
// Если хоть одно поле отличается(title, text, url, is_active) значит обновляем запись

// 2. Смотрим в таблицу tag_feature - существует ли запись с таким F и T, если да, значит обновление
// происходит самих полей, если таких записей нет, то мы добавляем запись, при этом, нам необходимо убедиться
// что передаваемая фича такая же, как в таблице у текущего баннера, иначе нарушится консистентность данных

// 3. Смотрим, что в таблице history_banner у записи с указанным ID поля title text url отличаются от обновляемых
// если да, то обновляем и инкрементируем номер версии, если нет, значит добавляли фичу и тэг к баннеру
// или активировали/деактивировали баннер

func (d dbase) UpdateBanner(ctx context.Context, bannerBody models.BannerBody, bannerID int) (bool, error) {

	var (
		freshOldBanner        models.BannerContent
		feature, tag, version int
		existsID              int
	)

	tx, err := d.db.Begin()
	if err != nil {
		return false, err
	}

	defer tx.Rollback()

	// Делаем проверку существования баннера
	row := tx.QueryRowContext(ctx, `SELECT banner_id
									FROM actual_banner
									WHERE banner_id = $1`, bannerID)
	if err = row.Scan(&existsID); err != nil {
		if sql.ErrNoRows == err {
			return false, nil
		}
		return false, err
	}

	// 1. Делаем обновления таблицы actual_banner
	_, err = tx.ExecContext(ctx, `UPDATE actual_banner
								SET title = $1,
								text = $2,
								url = $3,
								is_active = $4
								WHERE banner_id = $5`,
		bannerBody.Content.Title,
		bannerBody.Content.Text,
		bannerBody.Content.Url,
		bannerBody.Active,
		bannerID,
	)

	// Надо проверить обработку ошибки, когда записи не обновились, потому что нечего обновлять
	if err != nil {
		if sql.ErrNoRows != err {
			return false, err
		}
	}

	// 2. Делаем обновления таблицы tag_feature
	row = tx.QueryRowContext(ctx, `SELECT feature_id 
									FROM tag_feature
									WHERE banner_id = $1`, bannerID)
	if err = row.Scan(&feature); err != nil {
		return false, err
	}

	// Проверяем, что фича у обновляемого баннера совпадает с фичами, которые есть у баннера
	// чтобы не нарушить условия хранения баннеров
	if feature != int(bannerBody.FeatureID) {
		return false, errors.New("feature not comparable")
	}

	row = tx.QueryRowContext(ctx, `SELECT feature_id, tag_id
									FROM tag_feature
									WHERE banner_id = $1`, bannerID)
	if err = row.Scan(&feature, &tag); err != nil {
		return false, err
	}

	// Проверяем, что записи с tag и feature нет еще
	if tag != int(bannerBody.TagID) && feature != int(bannerBody.FeatureID) {
		_, err = tx.ExecContext(ctx, `INSERT INTO tag_feature
									(tag_id, feature_id, banner_id)
									VALUES($1, $2, $3)`,
			bannerBody.FeatureID,
			bannerBody.TagID,
			bannerID,
		)
		if err != nil {
			return false, err
		}
	}

	// 3. Делаем обновления таблицы history_banner
	row = tx.QueryRowContext(ctx, `SELECT title, text, url, version
									FROM history_banner
									WHERE banner_id = $1 ORDER BY VERSION DESC
									LIMIT 1`, bannerID)
	if err = row.Scan(&freshOldBanner.Title, &freshOldBanner.Text, &freshOldBanner.Url, &version); err != nil {
		return false, err
	}

	// Если поля совпадают значит обновляли флаг is_active или tag_id и feture_id и обновления не требуется
	if freshOldBanner.Text != bannerBody.Content.Text ||
		freshOldBanner.Title != bannerBody.Content.Title ||
		freshOldBanner.Url != bannerBody.Content.Url {

		version += 1
		_, err = tx.ExecContext(ctx, `INSERT INTO history_banner
									(banner_id, version, title, text, url)
									VALUES($1, $2, $3, $4, $5)`,
			bannerID,
			version,
			bannerBody.Content.Title,
			bannerBody.Content.Text,
			bannerBody.Content.Url,
		)
		if err != nil {
			return false, err
		}

	}

	err = tx.Commit()
	if err != nil {
		return false, err
	}

	return true, nil
}

// 1. Сделаем выборку из таблицы tag_feature, чтобы определить ID баннеров.
// 2. Выберем все записи из actual_banner по ID.
func (d dbase) GetBanners(ctx context.Context, queryParam models.Query) ([]models.ResponseBody, error) {

	banners := make([]models.ResponseBody, 0)

	tx, err := d.db.Begin()
	if err != nil {
		return []models.ResponseBody{}, err
	}

	defer tx.Rollback()

	// 1. Выборка из tag_feature
	rows, err := tx.QueryContext(ctx, `SELECT DISTINCT actual_banner.banner_id,
										actual_banner.title
										actual_banner.text
										actual_banner.url
										actual_banner.is_active
										FROM actual_banner
										INNER JOIN tag_feature
										ON actual_banner.banner_id = tag_feature.banner_id
										WHERE tag_feature.tag_id = $1
										OR tag_feature.feature_id = $2`,
		queryParam.TagID,
		queryParam.FeatureID,
	)
	if err != nil {
		return nil, err
	}

	for rows.Next() {

		var banner models.ResponseBody

		err = rows.Scan(&banner.BannerID, &banner.Content.Title, &banner.Content.Text, &banner.Content.Url, &banner.Active)
		if err != nil {
			return []models.ResponseBody{}, err
		}
		banner.FeatureID = uint32(queryParam.FeatureID)
		banners = append(banners, banner)
	}

	// проверяем на ошибки
	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return banners, nil
}

// Просто получение баннера по фиче и тэгу
func (d dbase) GetBanner(ctx context.Context, featureID, tagID int) (models.BannerContent, error) {
	var banner models.BannerContent

	row := d.db.QueryRowContext(ctx, `SELECT actual_banner.title
								actual_banner.text
								actual_banner.url
								FROM actual_banner
								INNER JOIN tag_feature
								ON actual_banner.banner_id = tag_feature.banner_id
								WHERE tag_feature.tag_id = $1
								AND tag_feature.feature_id = $2
								AND actual_banner.is_active = true`,
		tagID,
		featureID,
	)
	if err := row.Scan(&banner.Title, &banner.Text, &banner.Url); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.BannerContent{}, nil // Условимся, что если не нашли баннер, то ничего не возвращаем
		}
		return models.BannerContent{}, err
	}
	return models.BannerContent{}, nil
}

// Просто удаление из БД баннера
func (d dbase) DeleteBanner(ctx context.Context, bannerID int) error {
	tx, err := d.db.Begin()
	if err != nil {
		return err
	}

	defer tx.Rollback()

	// Удаляем из всех таблиц
	_, err = tx.ExecContext(ctx, `DELETE FROM tag_feature
									WHERE banner_id = $1`,
		bannerID,
	)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, `DELETE FROM actual_banner
									WHERE banner_id = $1`,
		bannerID,
	)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, `DELETE FROM history_banner
									WHERE banner_id = $1`,
		bannerID,
	)
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}
	return nil
}

func (d dbase) GetHistoryBanner(ctx context.Context, bannerID int) ([]models.BannerHistory, error) {

	banners := make([]models.BannerHistory, 0)

	rows, err := d.db.QueryContext(ctx, `SELECT banner_id, version, title text, url
											FROM history_banner
											WHERE banner_id = $1`,
		bannerID,
	)
	if err != nil {
		return nil, err
	}

	for rows.Next() {

		var banner models.BannerHistory

		err = rows.Scan(&banner.BannerID, &banner.Version, &banner.Title, &banner.Text, &banner.Url)
		if err != nil {
			return nil, err
		}

		banners = append(banners, banner)
	}

	// проверяем на ошибки
	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return banners, nil

}

func (d dbase) UpdateVersion(ctx context.Context, bannerVersion models.BannerHistory) error {
	_, err := d.db.ExecContext(ctx, `UPDATE actual_banner
									SET title = $1,
									text = $2,
									url = $3,
									WHERE banner_id = $4`,
		bannerVersion.Title,
		bannerVersion.Text,
		bannerVersion.Url,
		bannerVersion.BannerID,
	)
	if err != nil {
		return err
	}
	return nil
}
