// Copyright 2016 The Cockroach Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or
// implied. See the License for the specific language governing
// permissions and limitations under the License.

package apd

import "math/big"

var (
	bigOne  = big.NewInt(1)
	bigTwo  = big.NewInt(2)
	bigFive = big.NewInt(5)
	bigTen  = big.NewInt(10)

	decimalZero      = New(0, 0)
	decimalOneEighth = New(125, -3)
	decimalHalf      = New(5, -1)
	decimalOne       = New(1, 0)
	decimalTwo       = New(2, 0)
	decimalThree     = New(3, 0)
	decimalEight     = New(8, 0)

	decimalQuoC1 = makeConstWithPrecision(str48Div17)
	decimalQuoC2 = makeConstWithPrecision(str32Div17)

	decimalCbrtC1 = makeConst(strCbrtC1)
	decimalCbrtC2 = makeConst(strCbrtC2)
	decimalCbrtC3 = makeConst(strCbrtC3)

	// ln(10)
	decimalLn10 = makeConstWithPrecision(strLn10)
	// 1/ln(10)
	decimalInvLn10 = makeConstWithPrecision(strInvLn10)
)

func makeConst(strVal string) *Decimal {
	d := &Decimal{}
	_, _, err := d.SetString(strVal)
	if err != nil {
		panic(err)
	}
	return d
}

// constWithPrecision implements a look-up table for a constant, rounded-down to
// various precisions. The point is to avoid doing calculations with all the
// digits of the constant when a smaller precision is required.
type constWithPrecision struct {
	unrounded Decimal
	vals      []Decimal
}

func makeConstWithPrecision(strVal string) *constWithPrecision {
	c := &constWithPrecision{}
	if _, _, err := c.unrounded.SetString(strVal); err != nil {
		panic(err)
	}
	// The length of the string might be one higher than the available precision
	// (because of the decimal point), but that's ok.
	maxPrec := uint32(len(strVal))
	for p := uint32(1); p < maxPrec; p *= 2 {
		var d Decimal

		ctx := Context{
			Precision:   p,
			Rounding:    RoundHalfUp,
			MaxExponent: MaxExponent,
			MinExponent: MinExponent,
		}
		_, err := ctx.Round(&d, &c.unrounded)
		if err != nil {
			panic(err)
		}
		c.vals = append(c.vals, d)
	}
	return c
}

// get returns the given constant, rounded down to a precision at least as high
// as the given precision.
func (c *constWithPrecision) get(precision uint32) *Decimal {
	i := 0
	// Find the smallest precision available that's at least as high as precision,
	// i.e. Ceil[ log2(p) ] = 1 + Floor[ log2(p-1) ]
	if precision > 1 {
		precision--
		i++
	}
	for precision >= 16 {
		precision /= 16
		i += 4
	}
	for precision >= 2 {
		precision /= 2
		i++
	}
	if i >= len(c.vals) {
		return &c.unrounded
	}
	return &c.vals[i]
}

const strLn10 = "2.3025850929940456840179914546843642076011014886287729760333279009675726096773524802359972050895982983419677840422862486334095254650828067566662873690987816894829072083255546808437998948262331985283935053089653777326288461633662222876982198867465436674744042432743651550489343149393914796194044002221051017141748003688084012647080685567743216228355220114804663715659121373450747856947683463616792101806445070648000277502684916746550586856935673420670581136429224554405758925724208241314695689016758940256776311356919292033376587141660230105703089634572075440370847469940168269282808481184289314848524948644871927809676271275775397027668605952496716674183485704422507197965004714951050492214776567636938662976979522110718264549734772662425709429322582798502585509785265383207606726317164309505995087807523710333101197857547331541421808427543863591778117054309827482385045648019095610299291824318237525357709750539565187697510374970888692180205189339507238539205144634197265287286965110862571492198849978748873771345686209167058498078280597511938544450099781311469159346662410718466923101075984383191912922307925037472986509290098803919417026544168163357275557031515961135648465461908970428197633658369837163289821744073660091621778505417792763677311450417821376601110107310423978325218948988175979217986663943195239368559164471182467532456309125287783309636042629821530408745609277607266413547875766162629265682987049579549139549180492090694385807900327630179415031178668620924085379498612649334793548717374516758095370882810674524401058924449764796860751202757241818749893959716431055188481952883307466993178146349300003212003277656541304726218839705967944579434683432183953044148448037013057536742621536755798147704580314136377932362915601281853364984669422614652064599420729171193706024449293580370077189810973625332245483669885055282859661928050984471751985036666808749704969822732202448233430971691111368135884186965493237149969419796878030088504089796185987565798948364452120436982164152929878117429733325886079159125109671875109292484750239305726654462762009230687915181358034777012955936462984123664970233551745861955647724618577173693684046765770478743197805738532718109338834963388130699455693993461010907456160333122479493604553618491233330637047517248712763791409243983318101647378233796922656376820717069358463945316169494117018419381194054164494661112747128197058177832938417422314099300229115023621921867233372683856882735333719251034129307056325444266114297653883018223840910261985828884335874559604530045483707890525784731662837019533922310475275649981192287427897137157132283196410034221242100821806795252766898581809561192083917607210809199234615169525990994737827806481280587927319938934534153201859697110214075422827962982370689417647406422257572124553925261793736524344405605953365915391603125244801493132345724538795243890368392364505078817313597112381453237015084134911223243909276817247496079557991513639828810582857405380006533716555530141963322419180876210182049194926514838926922937079"

const strInvLn10 = "0.4342944819032518276511289189166050822943970058036665661144537831658646492088707747292249493384317483187061067447663037336416792871589639065692210646628122658521270865686703295933708696588266883311636077384905142844348666768646586085135561482123487653435434357317253835622281395603048646652366095539377356176323431916710991411597894962993512457934926357655469077671082419150479910989674900103277537653570270087328550951731440674697951899513594088040423931518868108402544654089797029863286828762624144013457043546132920600712605104028367125954846287707861998992326748439902348171535934551079475492552482577820679220140931468164467381030560475635720408883383209488996522717494541331791417640247407505788767860971099257547730046048656049515610057985741340272675201439247917970859047931285212493341197329877226463885350226083881626316463883553685501768460295286399391633510647555704050513182342988874882120643595023818902643317711537382203362634416478397146001858396093006317333986134035135741787144971453076492968331392399810608505734816169809280016199523523117237676561989228127013815804248715978344927215947562057179993483814031940166771520104787197582531617951490375597514246570736646439756863149325162498727994852637448791165959219701720662704559284657036462635675733575739369673994570909602526350957193468839951236811356428010958778313759442713049980643798750414472095974872674060160650105375287000491167867133309154761441005054775930890767885596533432190763128353570304854020979941614010807910607498871752495841461303867532086001324486392545573072842386175970677989354844570318359336523016027971626535726514428519866063768635338181954876389161343652374759465663921380736144503683797876824369028804493640496751871720614130731804417180216440993200651069696951247072666224570004229341407923361685302418860272411867806272570337552562870767696632173672454758133339263840130320038598899947332285703494195837691472090608812447825078736711573033931565625157907093245370450744326623349807143038059581776957944070042202545430531910888982754062263600601879152267477788232096025228766762416332296812464502577295040226623627536311798532153780883272326920785980990757434437367248710355853306546581653535157943990070326436222520010336980419843015524524173190520247212241110927324425302930200871037337504867498689117225672067268275246578790446735268575794059983346595878592624978725380185506389602375304294539963737367434680767515249986297676732404903363175488195323680087668648666069282082342536311304939972702858872849086258458687045569244548538607202497396631126372122497538854967981580284810494724140453341192674240839673061167234256843129624666246259542760677182858963306586513950932049023032806357536242804315480658368852257832901530787483141985929074121415344772165398214847619288406571345438798607895199435011532826457742311266817183284968697890904324421005272233475053141625981646457044538901148313760708445483457955728303866473638468537587172210685993933008378534367552699899185150879055911525282664"

const str48Div17 = "2.8235294117647058823529411764705882352941176470588235294117647058823529411764705882352941176470588235294117647058823529411764705882352941176470588235294117647058823529411764705882352941176470588235294117647058823529411764705882352941176470588235294117647058823529411764705882352941176470588235294117647058823529411764705882352941176470588235294117647058823529411764705882352941176470588235294117647058823529411764705882352941176470588235294117647058823529411764705882352941176470588235294117647058823529411764705882352941176470588235294117647058823529411764705882352941176470588235294117647058823529411764705882352941176470588235294117647058823529411764705882352941176470588235294117647058823529411764705882352941176470588235294117647058823529411764705882352941176470588235294117647058823529411764705882352941176470588235294117647058823529411764705882352941176470588235294117647058823529411764705882352941176470588235294117647058823529411764705882352941176470588235294117647058823529411764705882352941176470588235294117647058823529411764705882352941176470588235294117647058823529411764705882352941176470588235294117647058823529411764705882352941176470588235294117647058823529411764705882352941176470588235294117647058823529411764705882352941176470588235294117647058823529411764705882352941176470588235294117647058823529411764705882352941176470588235294117647058823529411764705882352941176470588235294117647058823529411764705882352941176470588235294117647058823529411764705882352941176470588235294117647058823529411764705882352941176470588235294117647058823529411764705882352941176470588235294117647058823529411764705882352941176470588235294117647058823529411764705882352941176470588235294117647058823529411764705882352941176470588235294117647058823529411764705882352941176470588235294117647058823529411764705882352941176470588235294117647058823529411764705882352941176470588235294117647058823529411764705882352941176470588235294117647058823529411764705882352941176470588235294117647058823529411764706"

const str32Div17 = "1.8823529411764705882352941176470588235294117647058823529411764705882352941176470588235294117647058823529411764705882352941176470588235294117647058823529411764705882352941176470588235294117647058823529411764705882352941176470588235294117647058823529411764705882352941176470588235294117647058823529411764705882352941176470588235294117647058823529411764705882352941176470588235294117647058823529411764705882352941176470588235294117647058823529411764705882352941176470588235294117647058823529411764705882352941176470588235294117647058823529411764705882352941176470588235294117647058823529411764705882352941176470588235294117647058823529411764705882352941176470588235294117647058823529411764705882352941176470588235294117647058823529411764705882352941176470588235294117647058823529411764705882352941176470588235294117647058823529411764705882352941176470588235294117647058823529411764705882352941176470588235294117647058823529411764705882352941176470588235294117647058823529411764705882352941176470588235294117647058823529411764705882352941176470588235294117647058823529411764705882352941176470588235294117647058823529411764705882352941176470588235294117647058823529411764705882352941176470588235294117647058823529411764705882352941176470588235294117647058823529411764705882352941176470588235294117647058823529411764705882352941176470588235294117647058823529411764705882352941176470588235294117647058823529411764705882352941176470588235294117647058823529411764705882352941176470588235294117647058823529411764705882352941176470588235294117647058823529411764705882352941176470588235294117647058823529411764705882352941176470588235294117647058823529411764705882352941176470588235294117647058823529411764705882352941176470588235294117647058823529411764705882352941176470588235294117647058823529411764705882352941176470588235294117647058823529411764705882352941176470588235294117647058823529411764705882352941176470588235294117647058823529411764705882352941176470588235294117647058823529411764705882352941176471"

const (
	// Cbrt uses a quadratic polynomial that approximates the cube root
	// of x when 0.125 <= x <= 1. This approximation is the starting point
	// of the convergence loop. Coefficients are from:
	// https://people.freebsd.org/~lstewart/references/apple_tr_kt32_cuberoot.pdf
	strCbrtC1 = "-0.46946116"
	strCbrtC2 = "1.072302"
	strCbrtC3 = "0.3812513"
)
