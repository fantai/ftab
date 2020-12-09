package cn

import (
	"fmt"
	"math/rand"

	"github.com/fantai/ftab/pkg/mock/helper"
)

var xing = []rune("赵钱孙李周吴郑王冯陈褚卫蒋沈韩杨朱秦尤许何吕施张孔曹严华金魏陶姜戚谢邹喻柏水窦章云苏潘葛奚范彭郎鲁韦昌马苗凤花方俞任袁柳酆鲍史唐费廉岑薛雷贺倪汤滕殷罗毕郝邬安常乐于时傅皮卞齐康伍余元卜顾孟平黄和穆萧尹姚邵湛汪祁毛禹狄米贝明臧计伏成戴谈宋茅庞熊纪舒屈项祝董梁杜阮蓝闵席季麻强贾路娄危江童颜郭梅盛林刁钟丘徐邱骆高夏蔡田樊胡凌霍虞万支柯昝管卢莫经房裘缪干解应宗丁宣贲邓单杭洪包诸左石崔吉钮龚程嵇邢滑裴陆荣翁荀羊於惠甄曲家封芮羿储靳汲邴糜松井段富巫乌焦巴弓牧隗山谷车侯宓蓬全郗班仰秋仲伊宫宁仇栾暴甘钭厉戎祖武符刘景詹束龙叶幸司韶郜黎蓟薄印宿白怀蒲台从鄂索咸籍赖卓蔺屠蒙池乔阴郁胥能苍双闻莘党翟谭贡劳逢逄姬申扶堵冉宰郦雍璩桑桂濮牛寿通边扈燕冀郏浦尚农温别庄晏柴瞿阎充慕连茹习宦艾鱼容向古易慎戈廖庚终暨居衡步都耿满弘匡国文寇广禄阙东欧殳沃利蔚越夔隆师巩厍聂晁勾敖融冷訾辛阚那简饶空曾毋沙乜养鞠须丰巢关蒯相查荆红游竺权逯盖益桓公万俟司马上官欧阳夏侯诸葛闻人东方赫连皇甫尉迟公羊澹台公冶宗政濮阳淳于单于太叔申屠公孙仲孙轩辕令狐钟离宇文长孙慕容鲜于闾丘司徒司空亓官司寇仉督子车颛孙端木巫马公西漆雕乐正壤驷公良拓拔夹谷宰父谷粱晋楚阎法汝鄢涂钦段干百里东郭南门呼延归海羊舌微生岳帅缑亢况后有琴梁丘左丘东门西门商牟佘佴伯赏南宫墨哈谯笪年爱阳佟")
var ming = []rune("莎锦黛青倩婷姣婉娴瑾颖露瑶怡婵雁蓓纨仪荷丹蓉眉君琴蕊薇菁梦岚苑婕馨瑗琰韵融园艺咏卿聪澜纯毓悦昭冰爽琬茗羽希宁欣飘育滢馥筠柔竹霭凝晓欢霄枫芸菲寒伊亚宜可姬舒影荔枝思丽秀娟英华慧巧美娜静淑惠珠翠雅芝玉萍红娥玲芬芳燕彩春菊勤珍贞莉兰凤洁梅琳素云莲真环雪荣爱妹霞香月莺媛艳瑞凡佳涛昌进林有坚和彪博诚先敬震振壮会群豪心邦承乐绍功松善厚庆磊民友裕河哲江超浩亮政谦亨奇固之轮翰朗伯宏言若鸣朋斌梁栋维启克伦翔旭鹏泽晨辰士以建家致树炎德行时泰盛雄琛钧冠策腾伟刚勇毅俊峰强军平保东文辉力明永健世广志义兴良海山仁波宁贵福生龙元全国胜学祥才发成康星光天达安岩中茂武新利清飞彬富顺信子杰楠榕风航弘")
var shuzi = []rune("0123456789")
var zimu = []rune("abcdefghijklmnopqrstuvwxyz")
var mobileHeader = []string{"139", "137", "135", "133", "131", "150", "153", "155", "157", "159", "188", "187", "189", "179", "170", "173", "175"}
var emailDomians = []string{"qq.com", "163.com", "126.com", "gmail.com"}

// Mocker is the mocker for china
type Mocker struct {
}

// IDCard implement interface
func (m *Mocker) IDCard() string {

	addr := rand.Intn(len(gb2260))
	if addr%2 != 0 {
		addr--
		if addr < 0 {
			addr = 0
		}
	}

	id := fmt.Sprintf(
		"%v%d%02d%02d%03d",
		gb2260[addr],
		1950+rand.Intn(50),
		rand.Intn(12),
		rand.Intn(29),
		rand.Intn(1000),
	)

	// last checksum
	iS := 0
	iW := []int{7, 9, 10, 5, 8, 4, 2, 1, 6, 3, 7, 9, 10, 5, 8, 4, 2}
	szVerCode := []string{"1", "0", "X", "9", "8", "7", "6", "5", "4", "3", "2"}
	for i, c := range id {
		iS = iS + int(c-'0')*iW[i]

	}
	return id + szVerCode[iS%11]
}

// Mobile implement interface
func (m *Mocker) Mobile() string {
	return mobileHeader[rand.Intn(len(mobileHeader))] + helper.RandString(shuzi, 8)
}

// EMail implement interface
func (m *Mocker) EMail() string {
	return fmt.Sprintf("%s%s@%s",
		helper.RandString(zimu, 6),
		helper.RandString(shuzi, 4),
		emailDomians[rand.Intn(len(emailDomians))])
}

// Name implement interface
func (m *Mocker) Name() string {
	return helper.RandString(xing, 1) + helper.RandString(ming, 2)
}
