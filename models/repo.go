package models

import (
	"fmt"
	"time"

	"github.com/dockercn/wharf/utils"
)

type Repository struct {
	UUID          string   `json:"UUID"`          //全局唯一的索引, LedisDB中 RepositoryList 保存全局所有的仓库名列表信息。 LedisDB 独立保存每个Repository信息到一个HASH，名字为{UUID}
	Repository    string   `json:"repository"`    //仓库名称 全局唯一，不可修改
	Namespace     string   `json:"namespace"`     //仓库所有者的名字
	NamespaceType bool     `json:"namespacetype"` // false 为普通用户，true为组织
	Organization  string   `json:"organization"`  //如果仓库属于一个team，那么在此记录team所属组织
	Tags          []string `json:"tags"`          //保存此仓库所有tag的对应UUID
	Starts        []string `json:"starts"`        //此仓库Start的UUID列表
	Comments      []string `json:"comments"`      //此仓库Comment的对应UUID列表
	Description   string   `json:"description"`   //保存 Markdown 格式
	JSON          string   `json:"json"`          //Docker 客户端上传的 Images 信息，JSON 格式。
	Dockerfile    string   `json:"dockerfile"`    //生产 Repository 的 Dockerfile 文件内容
	Agent         string   `json:"agent"`         //docker 命令产生的 agent 信息
	Links         string   `json:"links"`         //保存 JSON 的信息，保存官方库的 Link，产生 repository 库的 Git 库地址
	Size          int64    `json:"size"`          //仓库所有 Image 的大小 byte
	Uploaded      bool     `json:"uploaded"`      //上传完成标志
	Checksum      string   `json:"checksum"`      //
	Checksumed    bool     `json:"checksumed"`    //Checksum 检查标志
	Labels        string   `json:"labels"`        //用户设定的标签，和库的 Tag 是不一样
	Icon          string   `json:"icon"`          //
	Sign          string   `json:"sign"`          //
	Privated      bool     `json:"privated"`      //私有 Repository
	Clear         string   `json:"clear"`         //对 Repository 进行了杀毒，杀毒的结果和 status 等信息以 JSON 格式保存
	Cleared       bool     `json:"cleared"`       //对 Repository 是否进行了杀毒处理
	Encrypted     bool     `json:"encrypted"`     //是否加密
	Created       int64    `json:"created"`       //
	Updated       int64    `json:"updated"`       //
}

func (repository *Repository) Has(namespace, repositoryName string) (isHas bool, UUID []byte, err error) {
	UUID, err = GetUUID("repository", fmt.Sprintf("%s:%s", namespace, repositoryName))
	if err != nil {
		return false, nil, err
	}
	if len(UUID) <= 0 {
		return false, nil, nil
	}
	err = Get(repository, UUID)
	return true, UUID, err
}

func (repository *Repository) Save() (err error) {
	err = Save(repository, []byte(repository.UUID))
	if err != nil {
		return err
	}
	_, err = LedisDB.HSet([]byte(GLOBAL_REPOSITORY_INDEX), []byte(fmt.Sprintf("%s:%s", repository.Namespace, repository.Repository)), []byte(repository.UUID))
	if err != nil {
		return err
	}
	return nil
}
func (repository *Repository) Remove() (err error) {
	_, err = LedisDB.HSet([]byte(fmt.Sprintf("%s_remove", GLOBAL_REPOSITORY_INDEX)), []byte(fmt.Sprintf("%s:%s", repository.Namespace, repository.Repository)), []byte(repository.UUID))
	if err != nil {
		return err
	}
	_, err = LedisDB.HDel([]byte(GLOBAL_REPOSITORY_INDEX), []byte(fmt.Sprintf("%s:%s", repository.Namespace, repository.Repository)))
	if err != nil {
		return err
	}
	return nil
}

func (repo *Repository) Put(namespace, repository, json, agent string) error {
	isHas, _, err := repo.Has(namespace, repository)
	if err != nil {
		return err
	}

	if !isHas {
		repo.UUID = string(utils.GeneralKey(fmt.Sprintf("%s:%s", namespace, repository)))
		repo.Created = time.Now().Unix()
	}

	repo.Namespace = namespace
	repo.Repository = repository
	repo.JSON = json
	repo.Agent = agent
	repo.Updated = time.Now().Unix()

	repo.Checksumed = false
	repo.Uploaded = false

	err = repo.Save()

	if err != nil {
		return err
	}

	return nil
}

func (repo *Repository) PutTag(imageId, namespace, repository, tag string) error {

	isHas, _, err := repo.Has(namespace, repository)

	if err != nil {
		return err
	}
	if !isHas {
		return fmt.Errorf("没有在数据库中查询到要更新的 Repository 数据")
	}

	image := new(Image)
	isHas, _, err = image.Has(imageId)
	if err != nil {
		return err
	}
	if !isHas {
		return fmt.Errorf("没有查询到 tag 依赖的 image 数据")
	}

	nowTag := new(Tag)
	//isHas, _, err = nowTag.Has(namespace, repository, imageId, tag)
	//		if err != nil {
	//		return fmt.Errorf("查询 tag 数据异常")
	//	}

	//if !isHas {
	nowTag.UUID = string(fmt.Sprintf("%s:%s:%s", namespace, repository, tag))
	//	}

	nowTag.Name = tag
	nowTag.ImageId = imageId
	nowTag.Namespace = namespace
	nowTag.Repository = repository
	err = nowTag.Save()
	if err != nil {
		return fmt.Errorf("保存 tag 数据异常")
	}
	repo.Tags = append(repo.Tags, nowTag.UUID)
	fmt.Println("*********************************", repo.Tags)
	repo.Save()

	return nil
}

func (repo *Repository) PutImages(namespace, repository string) error {

	isHas, _, err := repo.Has(namespace, repository)

	if err != nil {
		return err
	}
	if !isHas {
		return fmt.Errorf("没有在数据库中查询到要更新的 Repository 数据")
	}

	//repo.Checksum
	repo.Checksumed = true
	repo.Uploaded = true
	repo.Updated = time.Now().Unix()

	return nil
}

type Start struct {
	UUID       string `json:"UUID"`       //全局唯一的索引
	User       string `json:"user"`       //用户UUID，代表哪个用户加的星
	Repository string `json:"repository"` //仓库UUID，代表给哪个仓库加的星
	Time       int64  `json:"time"`       //代表加星的时间
}
type Comment struct {
	UUID       string `json:"UUID"`       //全局唯一的索引
	Comment    string `json:"comment"`    //评论的内容 markdown 格式保存
	User       string `json:"user"`       //用户UUID，代表哪个用户进行的评论
	Repository string `json:"repository"` //仓库UUID，代表评论的哪个仓库
	Time       int64  `json:"time"`       //代表评论的时间
}
type Privilege struct {
	UUID       string `json:"UUID"`       //全局唯一的索引
	Privilege  bool   `json:"privilege"`  //true 为读写，false为只读
	Team       string `json:"team"`       //此权限所属Team的UUID
	Repository string `json:"repository"` //此权限对应的仓库UUID
}

func (privilege *Privilege) Get(UUID string) (err error) {
	err = Get(privilege, []byte(UUID))
	if err != nil {
		return err
	}
	return nil

}

type Tag struct {
	UUID       string
	Name       string //
	ImageId    string //
	Namespace  string
	Repository string
	Sign       string //
}

func (tag *Tag) Has(namespace, repository, imageName, tagName string) (isHas bool, UUID []byte, err error) {
	UUID, err = GetUUID("tag", fmt.Sprintf("%s:%s:%s:%s", namespace, repository, imageName, tagName))
	if err != nil {
		return false, nil, err
	}
	if len(UUID) <= 0 {
		return false, nil, nil
	}
	err = Get(tag, UUID)
	return true, UUID, err
}

func (tag *Tag) Save() (err error) {
	err = Save(tag, []byte(tag.UUID))
	if err != nil {
		return err
	}
	_, err = LedisDB.HSet([]byte(GLOBAL_TAG_INDEX), []byte(fmt.Sprintf("%s:%s:%s:%s", tag.Namespace, tag.Repository, tag.ImageId, tag.Name)), []byte(tag.UUID))
	if err != nil {
		return err
	}
	return nil
}

func (tag *Tag) GetByUUID(uuid string) (err error) {
	err = Get(tag, []byte(uuid))
	return err
}
