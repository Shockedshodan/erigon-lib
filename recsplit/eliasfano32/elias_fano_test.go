/*
   Copyright 2022 Erigon contributors

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package eliasfano32

import (
	"testing"

	"github.com/ledgerwatch/erigon-lib/common"
	"github.com/stretchr/testify/assert"
	"golang.org/x/exp/slices"
)

func TestEliasFano(t *testing.T) {
	offsets := []uint64{1, 4, 6, 8, 10, 14, 16, 19, 22, 34, 37, 39, 41, 43, 48, 51, 54, 58, 62}
	count := uint64(len(offsets))
	maxOffset := offsets[0]
	for _, offset := range offsets {
		if offset > maxOffset {
			maxOffset = offset
		}
	}
	ef := NewEliasFano(count, maxOffset)
	for _, offset := range offsets {
		ef.AddOffset(offset)
	}
	ef.Build()
	for i, offset := range offsets {
		offset1 := ef.Get(uint64(i))
		assert.Equal(t, offset, offset1, "offset")
	}
	v, ok := ef.Search(37)
	assert.True(t, ok, "search1")
	assert.Equal(t, uint64(37), v, "search1")
	v, ok = ef.Search(0)
	assert.True(t, ok, "search2")
	assert.Equal(t, uint64(1), v, "search2")
	_, ok = ef.Search(100)
	assert.False(t, ok, "search3")
	v, ok = ef.Search(11)
	assert.True(t, ok, "search4")
	assert.Equal(t, uint64(14), v, "search4")
}

func TestIterator(t *testing.T) {
	offsets := []uint64{1, 4, 6, 8, 10, 14, 16, 19, 22, 34, 37, 39, 41, 43, 48, 51, 54, 58, 62}
	count := uint64(len(offsets))
	maxOffset := offsets[0]
	for _, offset := range offsets {
		if offset > maxOffset {
			maxOffset = offset
		}
	}
	ef := NewEliasFano(count, maxOffset)
	for _, offset := range offsets {
		ef.AddOffset(offset)
	}
	ef.Build()
	efi := ef.Iterator()
	i := 0
	for efi.HasNext() {
		assert.Equal(t, offsets[i], efi.Next(), "iter")
		i++
	}
}

func TestEFSearch(t *testing.T) {
	h := common.MustDecodeHex("0000000000002b83000000000154db0f2e7ad414545219c6186493f9f61b70dd89a7de7a221ed924947c8a8a6aaad1528badb7165f9c75d86ab7fdf6f3d62fd0c0044600128820bbf0e24b43166104945c73d1c5196ab281679e81115268a3965b026ae8a395664aacb7010b3cb0c92fcfcc35e18cb3debcf5df83bfc1135ad05187219000334c39124d54d150441575d65da6a1c61a75d559e7dd83155aa8e3914b5e8aa9b0c362bbedc0052bbcb0d35057fd76e38f2bbfbcfefbf33f430e3b0c41841a6cd051c72cb4d4620b3aecbc130f4e3a0145d4524d1976186289f9569f7f30fe18249e79ea09eaaed6661b6eb906432cf3cc34a7fd78e494e3de7bf0eebf3f81053e8c71061b9c8c824a2bafacf34e3cf5188554534f79f515628d85761a6aa9c1179f7cf3e1d8a38f4a3afa68a5977eba6cb8e3befb70c413fb1cf4d067935ebae9d8abcf7e0412cc91471f7f0ca24a31c64c034e420a2d341555577185576596edf61d78e1e917a0800496a8628b2f5e19a8a0836edaebb0c532db6cbefa62dc31d145737db7e28c4f8e79f1c6877f3e0515f8f043147cf4e18729bf24a3cc36de80c312659559665e80020e48a08d399209e9a69c4a3b30c105af9c35de79ebadbaecd3cf4f7ffd11a4a0c20d5558c14721a28c62cc38e59c83104433d1f4d45559ede5d7628fa9f6da73d2adc79e84166ab861924ab6f966a497ce4aabb4d7eedbefc928a76c75e4b0075ffcf3d7639f3d0725fc004414820c52c830c63463ce4004156410575e8d751869a8ad36dc7bf0c527df7c1a8298639061d269e79e9d966aecb1e29a1bf0c1135facb3d863932d38e392533e3cf6e193ffbefd1b70b083106cb441092aa9fc324c33d14cc3524b2ebda4165f8011169b6dd4556760830e52f8e3914a46f9a7a18b9e2aedb5dd7aabf1c61c773c78e1883f1ebdf5db67c081085980a1c9269cf4428e39efd8e311484b45a5d75ea9bd465b6de6a957218636de68669a7462baa9a7c322cbacb328bf1cf3cc61ab8d7aeaaa4f7f7dfb0d5880810644bc01471c72bc22cd34e0e473114629a1b5565b9785369a790aa298228b354219e59e7f8e6a2ab5d79a7c32ca29bf1df7dc77633ebbedba4bcf3efcf22790420b30d871071e79cc62cb2db97c038e38e6f8f41350443d3699659c6dc75d77e6b1d8a28b4a5669e59590568a69a6c936ebacb4f0ce5b2fbe4627bdb4d3576b3d37dd97777efbefe5b7ef3efefacf408310433841061b6e3c724b2fbe6cd3cd39e8c054934e66d9761b6eb97568a28a6716ea68afbe0acb72cd36ef0c74f0c31fcf80155760e1c92fc0c853cf472199b4956ebb0997dc820c6258a5955756ca6db7deee5c74d24b43ae79e79e4bff00044d3821ca28a5c0f28d48233d855565988d869e7a25ba18a698be0adbecb3d0728b31d5557b7d39e699732f41065e80520c33ce5c33926bb0e9b65b70c2b5e75e9042fe0a2cb1c526bc30d043fbfdb7e189c3ce7efb2ed4f0051b9c80928c340625e414575d79f5156ce18db71e89259a08a39ba396ca6aaefdfe5b30d24a3b6d75dda69f8ebaf2eab3cf7f002fc4b083169558a2cd36dc3034d14b3efd44145e9dd9761b6eb9e916208d67ae09e7a396eecaabbcf6ee1bb1cc37e75cf4d172db3d3ae9a59baefcf30a2c7041061b6cf1451863b4f28a2cb378138f3fffc024134d49c145d75d94e1979f7efbf1d79f7fff8d98229351d6c9a7abb19a7b2eba03ab0c73cc38bf4d37de7cffed7af9e73b9081124dd0e1c71f8094e3ce4d3b81951668a7c9d7228c35862926abb0daba6bafbfc25c33ce3a132df8e084c7fe3efcf19f00430c5780d1871fa2bc828d36dd90f310442d3155d5557e4d065b6cb2e1a75f7f2a76f925998a8a4aeaaadd3e0c31c5198b5cf2d693bbfefaede3935ffe053f24b104136ff0f1c728e738345146201565145376e595d966a475e7dd773212d9a4937e0a5a68a2ca424b6db6f6de8b2fc9400f8d74d8784f4e79e596f72e7cf7e24b80c114545461c5158d50728926c680238e39047d049248449d85965a9149e65a6cbc09675e7a0232d82084391259a4976dd26967a6bd127bacb2d1062cf0c00d073d74d153535db5d59d473fbdf5def7e003104cd881c7239d80424c31de94738e3f1f9574124b32cd45575d82adc69a76eed9875f861a8eb822996dba69279eca2e9badb6186bcc71ce6cb71d77e4bbfb3e3cfaf2f75fc00a4c4021c82084d0720d36f3d0139040303d855556630d869a6ab5d9e6de7b000e3822892cbad865a08536baabc21057dc71c829af1cf8f2cc4b2fc31b72cce10721867053d34e3d1d05555455f9d51968a399c7a1871f0eaa28a38e423bb4d14a473d39e5955b7e79f02fc450c30f64d8a1072188c8720b38e188538e430f41e453565ac945d76391cd769b72cbb9079f82197e08e2924ce6a9e7aaaece6aabbbefd26b2fc925e7ac73d79c77ee79f7deff0f800436e4c043197e0022882ac414634c3b114934514d3bf1d4135c8c3916996becb9f7de8335da98239676dea927a7a18e4a6ab2cd422bedc109330cf1cb30cb9cf3dd7af72d38efbdfbfe7bfdf6df6fc516800cf2c82bb0c4624d3df7e033104b2d35b5565d7ae5065e7aecadc8628b2ebec8e7a08452cb74d34e3f0df5d68723fe38f4d14b0f81155760010a2cebb0f30e4d5a6de5d56ddb71d79d77df8107e18c36eee8a7a0852aaaeab5da72bb76dc74db9dbcf2cb334f82092bb0b08a2bb0c8a2514e3aedc45461861dc61a6ebaf54621861b7eb8e69b72d689ecbaecbedbb3d147235db7dd7dff0dbdf4d35f2fc514545411082ebbf0928e3aecc4e3545453619558638f4966dc7bf0c587618f416649a6a8a3925aaaa9a7a27ab1cf41170d77dc72d7adfcf2cc3fff000411d070461a6ac012cc30c51c5391451725351861872196dc72ccb5a7e08d38e6a865a08526ea2bb1c52a8befc720878c35e187278e3cf3f0c72ffffcf49300871c76acd2cd37fe0cc4d24b5d81455965c4210760800726a8238f67a239e8a1a8a6ea2db80b336c33ce58677d37dea4a34efdf6e0936f80020f5470861a6fa4a20a36d968f3104439ed74d45a9359265a78e28d475e791c7e58a28956ae19a79cc46acbedb81867bc71c851f7edf7dfb6ebbefbef4314710413df88830e3b1f412595629769171e79ebb597a1926fc219a7a8a7a6daecbe00833c72d45377adf9e6a1a7eefcf3f9a7b042136188f1461ca040438d36ff146554524a7d05566492d1565b6ebcfd07e08004e6b8238f5cfe1928a1a02edb6cb4ffa6acb2cb30e76c77de7dbf4efbede497df420c5a1c828822bbf0d28b2fbf4024d14417b1e5165c71e5c59767c61d875c72084228e1843f16f9669d79c21aabade1a6ebaebc1773acf4d26dcbfd38e49a7b0e7ae8efd36f7ffefdbf10830c38a4a1c61aaa8c630e3a0d3d149144307945986187c9365b75d6f1d7df7f019678228a5b629aa9b0c6da8befbe00935cb2c95f9faefaf505343081145300224b2db67c0492482641255555597d069a6fd98967de79eaa5a8628b4c021aa8a086e6aa6bafe9da7b2fbe2a9f9db6da6be39fbffefbf36f0412491c02092fbea833cf45180155d4524c8d465a69c9a5c79e861b9a7966a69ae6aa2bb0e10e5cf0c12a3f5df5d5588f9d77efbe1b8f3c061c80504213861c824822b2dc928d37fe10f4124c31e584d65a8c4d361b6de9b9c7e187211689649d77e6b9a7a7c92acb2cc310478cf3d662939df6db925bae7aecd86fffff000724e0400e4724e1441d7694a24a35d8d0c3904418d594555a6ec165d768a495469d76de11a8a08b33c6d929a9a59a7a2aaae816ec31c82a4b6d35d6b1cf7ebdf6e4477081061f84e0441e7aec114a29aab0b2cf3f038184945351cd059a68a8a58621871d7a28e2882ebec8679fab2abb2cb3010f7cb0cb36e32c77dd7c633e3ae9adc7debfffff7be0c3124e58a1c51f8720928822bf04334c400319d412565ee9b5175fda7907de79ebb1c7229ab3d68aabbaebb26bf1d27dfb0d38eeb9eb4e7e071e7c00821e7cf4114830c210734c430e4164514e3bf5349461898116da72cd41675d8b2ec6a823a28a32dae8b1c826ab6cc5196ffcf1d866a3dd76efbeff0e7cfcf2cf8fbf2187209288228b30d28d44145934925557fd95586db7e556e28d3afa48e491851aaaebaebeb6fb70c41473dcb4d376fb0d7becb4839fbefa114870030e39b84108228a9cb24a2cb4e8f3cf4020159514535f2dc65863ace9c61b70e565d8a187367e19669996126becb1ee4e4cb1c53a072d34d16bcf4db9e5b5b3dfbefb0a4830030d6ed891c71e7cd8b34f3f2a95b5165b9169169f7cf35979659d77fa3968a188c24a6db5d6de6b31c61b033d34d14bff3db8e18763effdf7efe7bf3f071e083144196928b2482cb39c030f46190595145d76f555986cb599c75e7bf2d9e7e28b33ee38679d8a529a2bafbf16bbaebbefdadbb6de7bf3fdbaecb35f6ffffdf96b80430e3ae4b148238e0c938e3aeb1824134d35e5f453596e496619669f89f65c7dfcfd67e08c34d62824917eca7a2baee9ce5bb2c927674df6d9696bfe79eebd8b0f410415dc1086186448320925b6f8f24e3cf43c44514d3759a5955f80b1d69a6bafc1b61d77e08168e289425e8965965a928a6aaaacca5b2fbefb0a4c70c15133fe38e4bbd78f7ffe03bcf0841457d8718729b4e4024e38fa0064524a4191b5165b755d865966ba15671c72fd79f861884a36e9a49e996eda29a8c89e8baebafb862cf2c94937edb4d4839b7e7aeacd57dfbfffff9f81c61a715c828927a3e4a3cf3efcd874134e4be1f518649149361965a531771f7e5f8649a699ac9e4bb0c11677ec31c848272d36d96ebf0d77dcb1d36efbedff031080000310504013514831451a861c124927a1dc924b32d558734d38e494830e50472195d457659945d76292b1d6da79eaadb71f8017629861872fc2e8659884167a28a3a2e29aabaedd821beebb1b73dcb1c7361b9df4d278e7adf7deadbf1ebbefe29b7f3efcfda7b0c20b67ace1861cb7ecf20b31c6cc44534d3619865862be11479e79f0d988638e6b62ba69a7a37afb6dc445c74df7ddb2434f3df6092ce0820c3af410c41abbf0728c32f4dc838f485c79f5555f924d465965062298e083171669649d9c76eae9afc7925b2ec3332fcd74d34eb72e3bedf3e7cf42104734b1871fa598720a34f1d473134e39f9f4135073d9755768ca39071d7c001668a08c5c967926a2abc62a6bb4f1da8bafbe3af7ec33d09fc32e3bedd0536f7df637ecc0830f831c62cb2efcf4e3cf3f001155145350c1465b6db98577628a2c3219e5945d7e0969a4bd061b6db5db720b6ec415831d36e4c20f6fbcf8e3935fbef9e73b60810c5458c1c7269c78628a34d5dce38f4107e194935a6bb5059768d86de7dd85190a49249472d27967a59612fb2cc0010f9c70ce5f836db6e79f9fceba020b30d08018639c818624935492893aebb0f3ce53504525d5545455351a6aa9b1961d871d7af821924ee6e9e7b1cd4e6c31ca2cbb7cf3d86497edf8ecb687df3e0631c8c0431d76dc414a32ebc4c38f3f02d9d45351492d8597679f81665c79f3d5a7618827a2b8e5a4baf2eaebbbf6de9bafc94c3f0d75db9253febaecb7e7ce7bfb1b7830821372d4a10724915892c931c928c30f492599e4525b6ec155996ebbf1765e8c33d668a3bcf4da7b6fe4d0473f3df7de835f3e071d7830862ae9a8b30e5046b5e5165d7fc5369b6decbdb7228b66229aa8a79f4e4bedb501d78cb3ce80072ef8e084dfdebdf8e33b00010d3aa0a1061b8e18738c32f0ec331249449985165b72e1d5976ab189371e813f0219a490433e2a2cb1c9f21b30ce3c5f8d75d65a6fddb9e7e87f408209567021c61867e0b24b2fc1d073cf3efde4b4934f44b1d69a6baf0118a080095a79259685da7a2baeb93a1cb1c4387f8db6da6bdb4dbae9d2afcfbefb33e8c0430f42b4e2ca2bb034e41044157934d257623d76196ebb01c75e7b072218228926666928a3ac525bedbbf0e2ab6fca4fcb3d37dfbaefcebbfe2aacc042178b30f288269b70e2893efc78045249269d849260a599861a75d55937e08a2cb648639b6ea29aaab3cf424bedc311bb4cb5d75f67deb9f7df836ffe0518381105228cf822cc3af04ca4d148524d45955687bd065b6cff1d8820831356e8a6a4996aca2cbdfcf60bb2ca32cf2c76dab5db7ebbff441461c4116848328926bbfc424e3a2635e5d4537bf9059866d251579d762bb2d8a28b5d4e6ae9a5c722cbecc94d3f1d35e2b9ebbe3bfbf23f00c10f421061c4137d54128a28ac78034e43162195d4537efd055860b6dd76dd810e4628618f431ee9e7aab0c66a2bc0041f9cb0ce3cfbacb6e6a09b8f3e04114830811e7cf8114830c414635147239164d46088f9f69b791666c8a193504e692597bdfa8a2db8e5128cb0cc36e7ac73d7955b7e79f5dcff0f4008451871441a79e8b1472cb57c034e411f855492555d79a5d763b9e9869c74d55d47a08c335ad925a28a2e1ae9aebdfa3aaec312532c73e393cf4f3f155f8011862831d154934e3e09851460cb35371d767cf63968a1a9aaba2aab1e7f0c72c895d7be3befc22bbf020b2db8e0c927a0a0f38e4b30e9b51a6bae0dc8a0830f4228e79c76de396eb9e7aacb36dcb0cb6ec1051b7440821967ac428e3a35ddb41364d769d7dda18826aa28ba33d74cf9edbaf70e41066db8f1862ef5e0d312576199055964b6e1e61b71c6bd472394524eb9e8a4945afab0c53cff6c78e38f4b6e79fffeff0f40165b7811462db6f0e20b39239574524c32cd4413618519765877df81171e91451a79a4a7a222abacbd2dbbfc72d5581b8eb8f6df834fbe0413f03084135024c2483fff0834d0565c75e55569d049371d8147229964a49666aae9bbf5de9b6fc967a39db6da6bb30dfaf1e5a3bfbe071f843082188724b2482fbe00b34e4e3cf904546fbf05271c86492ab9e4a393de8b6fbefaeeabb3d65b073e38f9e59b7f7e09269c500628a28cd2cf4414556491596bb505da77fd011820955b7209abbefbe6ac33e497639e7cf3041470400269a831071d800c528b2df1c8330f3d2ec1845556acb9065b78e28d475e864826a9e49c9c7afaa9b4ff0a5cb0c1385b7d35d687b3debaebe8a7afbefb1e7c00420876e091072bea74e4d14776e18598629769b6d97a3bf668e8a1c92adbb2cb38ebccb8e3ebb71fbffc341c824822c30024d0403aa9d5965ba77d4761851866f8a188556619a8a0a626ab6cb4dc32dcb0c334f7ecf3cf435b7e39e69c8f5ffef90c7c0042085a08428821ba483451451b99765a6aedf5e7df7f536ab9e5a5d8862b2ec85d936d76eaad1f8ffcf5116090411169a80149259e4c63cd36ce4d479d75db8537638e842ecaa8b528b71c33cd36ebddb7e0cf674f01061a60414b2db6dc12cf3df9ec54175ea981271e7920ba28e38c78d25aabadb7ba7befbe45271d78e1bd0b1fbffc1ca0a0821250f8f1872aad3c13cd3d001da410433b9155975d77e195576df0dd87628a4c56a9e79fa08e4aaaabf3ea7b32cb5867adf5d686336e3beedaafcffefb0f4080420a504c41471da194624a3929a9345659ac45379d750b32d8e09179863aaaaab9b20bafbc1a6fdc74d4a08baefaeaf0c72ffffc231ca104138da4c34e3bf6e4a3cf4d3e0d45d460bbf1d65b7d29ba18239fb5deaaeeba14a3bcf2cb632bbe38e4ff0720c0010b505081155880220a37dfdca45350430d461874d269a76188226ef9e5b0c4e6bb6fcb44176df4d16bc78dbaebdf9faffe061c74e081238f4432c93aed3c14d14825ad845560820d56db7aedc167a291482e3928adb6f20a6cb1ff02ac70c33e5b7e39e6bd07df8107229480461a6aac618b2e008d94d25a72ed169c70011669e49e9ef2da2bb1c8f65bb0cd3aa3adf9e7bd831fbef8e393cf84135314c2ca2bb5b493914721996458628d45661966cad1875f7e253a7965969a721a6cb1eaae1b6fc82213cd36dc75437efcf2f2d7bf3f0f410841871da49412ce38e9a8f3ce4651c115975ca28d461a6aeba5a862996d4a4a69a6d1be0b6fbcf25eccb1d579eb0db8f2cb33dfbcf3124c9041128310120c31faec53524c34e1e4d34f6f19a71c73d3fd67e081441af927a0810a0aecbe176b9c36df8313aef8f2082470810657608149269a6c724d3f031db4d04c5a6d0596618f4596db6ee28d879e7a2cb6e8e28ba08a7a2aaa4627edf4d391632efdf4f3e77f030e3938024924b1f8f28b30f5f083524a5969e5d7608ca9069b6de399a71e812ec648639571de89a7acd466ab6dc123f3dc33d06dcf4d37e59d7f0e7ae8edd35f7f06535051051d800ce24929e7a0938e3aeb4065575e821d96d86cda79071e822dca38639b9e925aaab61867ac71c750bf0d37dd9b874e7ae9ac3f20c10827dc81471e7aecc1872aaec40312494cc155d75d7cb1269b6dbaf1d7df8829460928a18bd29a2bb7dd7a1b32d3510b6e78e38f0fbfbcfefbd730061a6b5c124c31c728438d44145df4d35d79e9a51b7efcf977a090461ed968a4996afa6bbeff02ac74e0820fae7cf3d037308115914cc20937e14874134f3fe9b5d772d049a75d77e599f72188555ac9e7a7a9aa1aacb0c3128beeba187ffc33de9c7f2ebaf3cf5bcf80044c34e1c4135034024925c634c30e3c14e54454514951861967c1f917e0810c3688219a6ecae9e8b2d152cbf0c3142fcd74d399432f3df6156cc081086390918823ad64b38d410bb5149450431155d65aacb5065b76db7d179f7e2fc218e39f8932fa28adb9f28aedb7010ffc31d0435f8d75d65a6fadf8eacd47cfbefb010830800c56783146196b3c62c925b460e3cd3d186d34124e431d25d75c9559761966cd4137dd7a0f4248218b8826aae8a28c362aeeb8edbe0bb1c43cfb4c74d169ab7d3beecf477ffdfe197020021451483145219e7c524a350415b4504313ed84565a70c995976cc011971c7effb9482396591a6a29a6cd4e8badb602a39cb2ca4d439d35d7a5a31ebcf1e9b72f3f0619dc60451765cc91871f83bc320b2dd7d4938f3e2311b6d87bf6c56a2baec4169b6cc97dfb8dbdf624008144124c48a1ce3aefcc8497649665069a7b218e98e298aaaeca2abb22b73c33cd76effd37ecd97b4fbe08239050821559ec114827a4a4028b3aeb3cf4524c5469e515659af9761d76e49587638e3b0269e79da6b2eaeab40327acf0c229abbcf2d7b1cf8efbff062420010537643106197f00d28b2fc414a3cf3e410d75945771c965175ecf41179d7419b228e38c38ea58a79d7cfa096aa8cf461bb0c005931cb3cc53bfcdb8e69e83be7aebd7c7bf3fff2ebc50c821a7a4024b2cf518c59863d561d75d888c36eae8a390464a71c51c7f7c36da6dc36d77f1104c70011aa8ac020b40021994d0524c15a61870f2cd67df7e0316486491461ec967b1c82aebacb4041fec70d350b7edf8e39003affefaf4efffbf00369c91861ba85cd3cd3724c15557669ab5e7de7bf0c557269a69caf9aabaebdadbb3cf3fe73d3bedb57fefc30f406c220a29a5a412d0400521f4925a6b61065a72ca2dd79e851a8288229452668968a28b9abbaebbefde4b32d4514b2df7e28c670ebaf1c7432f7f071e3c2105157708720822ad4423cd34d460e4d1473c7525165988a1e6da6bd375379e820bba58e38d391e8a68a2b762bb2dbefc92acf2ca4d4f6df8e18bc72efbecdd036000020c3c5088218720720c39e8b4538f3dfb90f413505f99c55863925596d9660722b8e08c545ea9a59da08a6aeaa9af8a8baec033d35cb3cd57679df8e293d7eefbefff0310c0073dfc80861a6e18d288249518730f3ef9e8f3545457c1f51760872dd65c74d66547e185380a89649250463929a59a6e8aeebaefc21bb2c84e3f1dbbec0310c044134e18528c31ccc8d3934f3f011594507261e65968cff13720812a5ac96597aac26aebadc1ba4b6fbd24975cb4d179ef3d3ae9c61f8f7cf20284a0c2114dd0d1871f8414424d35d65c33d04829ad3415555d29b61863ba61875f7e2dc668238e6636fa6aacbf0edbaebc0d3b6c31cf810b3ef8edd55b2f80012aac10c30d3b9461462393b0f24a35dbe0d3cf47200d75545cc51d979c8b8e3e0aa9b4dc824b6eb9945fbe79e72cc440430d7ec0324b2dbb9c834e432ee1a4d34e4551569965c7ade7de7b1f8e68e2892dd219aaa8a3229becbaf56edcf1c728072ef8e0845f9fbdf604549081065a6cc18927cecc53cf3d1e011594506ad1555761a9b916db73fdf9f71f8001ca18a59459ee7969a6a0862a2ab8e24a5cf1cb4d636df7dd78972e3cf3d0633fc1054534118820a7bc128d34fb10d491474a2d35175d7b1156586bd26957a08127baf8e29373eab9e7b1fbf2dbafbfff66adf5d661837e3beeb9ebbe3bfc1f9cc0c4136ed881c726bc14730c3a0b4124d1454121c5565c764d765b6edd957720825a6ec965975eb2ea2aacb9fe0bf0c00a2fcd34d555871db9e4b90b2ffef8104420010d536871871eaebc024b2cb28ca3904334f144d6599351b65b6fdf29e8208437fe28e49049fa19aaa8a32e1bafbcf702fc34d4564b3eb9efbfbf1f3ffd23a4b002137c18a20b31ce5063cd454331e514549b7d261a69f4d567df7d4e4a4925a1976eeaecb3fc062c30c957771db6d86a535eb9e5c9e7cf7fff34f840061aabd4724b37145564d14e5d99a51667d57167e18535f218a4914a26da68afc4e2db6fcd36c76db7deb4efbe7efc5a6cc18524e3a0d34e4b3af904545dc011671c82421a79e4a5a09a8aaaaa24976cf2c9c41f8ffcf210587081158d50720926fb0034104d6bc135175dd04527dd74d46d98a4927f1eba6bb5e0968beec7401f9d74dd7ea39efaf8fa0f60000d61a091c62ae390534e441c75e5d55767cd461b74d64d58e1965c4e4aa9b8e9ca6befbd6fc31db7dccf53cf7fff33e0e00310a19c928a3517793412496091351965c9293720813d06f9269ca2aaba2aabe59adbf1c7514f6df8e1b9f7ce7efb36e8c0830f75b4f20a2cf09c94934e4d0526d860b8e9d69e7b32ce58a79d77e299a79ef7ea0b73cc7bf3ddb7dfca33df7c0035d870c31ee9a8b34e3b37e194934ebbf1765c7e39ea28649c7bfa2968a1061f8c70c274dbbdb7e9dc7b4fbe071f806045259b78620a2adb60f45148669de61a6c1866f821889a6eca69a7ff067c30c27afb5db8e1f5db7f3ffe39e8200411d760d3cd3741095514528111a6986cbfb1179f8f3f0219a49073d219ecb14523adf4d372cf1d3aeae463b001085760a1c9269c74e28936150525d4517efd055860ca2198a0838d462a29a5ca3e0cf1c48e433e39e5d5639081061b70d0c1168414624b2ffcf813504b6191959666b7e9d6db7d37e6c8639db5da7a2baef9fa1b30cd75efed37ebadbbcebdf726a8d0c216befc024c30c20044904e41e5c5d774d45567dd75d81589e4a28d526badc20b074db4df80a3cefd00041460c00108b4114728a45ca30d4129a9b4124b2df91598608785271e79e8a9b8228b55462ae9a4c0ca4bafbd14d79cb3ce737f2ebae9d91bb0c0105eb411071d7660b30d430d1da514537b21e6986ebb25d8e083121e9bf0c20ddfbcb6db72abbe3aebad8f51c6196a7c034e38e2dca4134f3e91671e7ae9e19866a083464ae9b4d726acf0c245e7ad37dfac135080010b1891c71e7fe4b28b3df7cc44534d3fe595da6ab091571e8723c658239863763aadb5da525cb1c552330e79f0c5330041134e24d28831cd28e45044176964d76ab1c9169f91472299a4928b428aecb2ff164c74d28417de3bf0ec779004135160e1cd37e03834115265ad45176db9b997628b33dea828a39022bb6fbf00bffc75dd78033e7bedb7b32f030d374c424925965c020d41061d6417639251f61968d05dc760830e1e5968b0c37a8b6ebb31d76cf6da9767eefefb1e809048249350920c420c35c45368a28d465a69a671c7a188236e29e9aebd42cbedc0042bbc73d06cc7adfaeaacabaf01096cbc41071e82cc82ce3af390c494545575361d75d5896822975ea249a9aebd264b6fbdf6d29d37dfca57ef800448247108228d60020c31f608541358618985996cb4d177df8928925966a59632dbacc108472df5eaadcf5e7befbfe73fc30e3d80420a2ad9dcc3cf403295a5165c9d51579d75063238618516ea1928a186f6daafbf0103ae78e38f4b300105155830441146fc110e39e5a84351454fad75d767a19da6dd76082e58269a69b6e92ab7dd829b31d34d3bdd78e49363ee3efdf8f3afc5165c7421cb32d14ca3d0431455f4d864986da61e8b32cef8a7a0882e1ab0c00313ac36dc73d78d3df8e33700430c32cc508a2baf80435451461d85945e7d99a61a78115a9861924a168ae8bcf4eedbefcf8e3f0e79e4924f0efef8e6832002208108320821c0f453934d3705155879e6c1775f7e4a2e49e8a1cf7eec33d046336df8e188fb9f40124b48b10a2b177584924a67b9059b78e7b967e2892bd278269a9c8eeaedb71267fc35d8618b6dfcf1cc5befc20b4c5871c8228dfc020f3df7f47354524b5155d870c415b7628b30e2a8279f7d061a6cc0032b2cf8e085375eb9e5975f6fc2092890b20a2bafe0d38f4d39edc5575f7e858761861c5ed9e5979d3a5bedb5da661cf5d4573f6e79faeaafcf7efbee5bf145228b1463ce39e8a4a3d146490dc69a6baff5e7df7f00f618a79c73ce7a2bc0013f2cf1c44e2b0eb9e4c1f72f40010630f184147eecf20b31e9fc23505146edf5d760b4e9c75f8005327824939162aa2dc5155b1cf5e39027ef7df9e69f808217607852ca35da889452575e5d065a68ccc1971f8c319ef966a8a3aedb2ebcf3fe6c34d27f3bff3cf4027ca04218634cd28c33d17c045248228dd49967a0a1e62494526a59a9a597622a30c108477c37e8a18b3ebafbf0a70083176154a28936dc98c38e492d493515558da107df7cf7fdd827a0832adaebbc104b3c75d65c7b0d3af7df77d0451863a0914d3ffefc03504042c9b5d76ec38d48a28b38e658a7a9acbaba30c61a878c72d65d535e39f1c937cf84135454e1ca2bb0c4b2124c357df5d864c519379e7af18978e59764429a29c0011fdc70d2540baefaebb69f1f84114b4861c516a44c934d3723558555618e15779c73d865d7a0834416c966aef9eacbafcd37e34cb6e9b0cf5efbee1c74e0c107851c924824993403cd360825a4905178e9351a6ce189e7e1874c3a09a5a6a27e1b2ebb2193fcf5eaacb71efffe35dca0831a9f84424a3a19a9c4524c37c165576aaff1276098628e496699976e8a6ebaf4c6ac37dfb9ef8e7cf2cb332fc30c35d8d08c33cf407394524b3105d9649559965f7fff05a8a38f410aa9ebaebcf65a73ce3aefacf9e8de7fefc20b30c430c9269d7872cf4518653415575d7d251970c10937e08d38eac829bae9aabb2ebbedba3b35d75e67ae3df8e1c310830c3390610927a0586310430d0115d7649479869d77df558864924a4a4ae9b2cd2edc70c55fa31dbae8efd78f3fff03cc51471eb0087490420d5565d55581f9261c72f2b5e8628d39fe38e7acb4d66aabc5186fec31e79d874e3afef9ebbf7f1d76ecd1ca39e8b053935968b5b51a800222a820934d3af9a4afbffa1bf0cf40ff0df8ecb4efdebbf0c59f80420b30445289269b380491454221a6d863914d475d7500f208e49d79eaba2bbd165f8c71c64723de38e4ddaf10030d3bf8c0461ba5a4520e4928b9b4177bedb9272292546a19e6989666abedb60433acf3ce6fdb0d7bedff03e004155954928927d44454114775fd359861bbc1371f7d1f7af925999b72daa9a79f6aecf1c859ebedb7e083e3df3f0001ac214730c430d490430ffd459861be0158a0814c36e9e49350daaaebaec1beacf6da6c6beefcf3d25bd0c107216cc1c5218cf0e24f4003cda456638e3d06d96fc11957e08237e698269baaaeca6aabe0860bf1c417e7adf9e7a1e31f00104328210b31c530334d44173d05d55f84e5365c7df9b548a39c73968aeaaaf8e6dbefbf39bf3d37f1cf031184104310414a29b074139144154dd59967a2a9261b6deb89a82299659a99a9a6cf662c72c92dbb7c76e59daf7fbf0b37e02044218e3cd24c3f082534d45f965de619692fc6f827a0b4d65af0c1104b8c71c78a2fce78e30628e0c0035010a2c822e298a34e3b6391559659d761e79d85669ec9669b79d66aebadf0969c72d459434ef9f5dc8fbf3efb2378418625975093cd3629b5a4935b739d265e79ebc918669b72ceb927aff0c6ab2fcc31a74d77deb9ef2ebffd33d0a0c321c928234d35db3c2494514b0546db6db9a997e28a5b7279679e7ec64aafbd2ab77cb8e2904b6efdf5d847c08414872c62ce39e8a4a3ce4721651556669b51c7dd851866a8619969764a2ab7e066fc71ca596f3db6ebafc38e420b3480b10928a29802524926a554135badbdc69b820b1659269cb4da8a6fbe26a73cb6e69b73ce7efbeebf0f3f096088514629f0c423cf3cf4d433925ba089865a76da6dc79d8b69b22967ace7aa2b6fbdfbe2dc38e49927bfffff081c81871e84b0828e3af4e4c3d24b3a2dd6186fc7c5271f8b36fe08e49daace5a2bae24b7ecb2cc3793cdb6e8a3539fbd010bc8b08729a7a0920a3719917412558a49861967a161979d8724828929adbf125b6fc41773ecf4d37b0beebaedeac74f8209820c4248218608334f3dfc6ce5d557668d56da69ae8118e288298e4a6aa9a96ecc71c71e5b7db5d661d77e3beef063b001075030f20824b5d463cf3d33fd64d45148a1c6da7aecadc8e28b32b6e9e6a69c1a7c30c209f7fc33d04373dffdf7e34311451a6bf0b18a2bb2f8b3124b3121a6d862902927e0800ad259e79d781a8ae8a2c4a6fb2ebc19ebec73d09163ce79e7d0331041125430128c30e4b0030f3d2fb115976baf81d7e18739ee98e79e80f6ea6bbf05073d78e18c576e39f6e28b4082095f90720a2ae9d44392496bb1d5965ba775f7e08723a2186399667eaa6cb31c931cf7dcce5baffdf61e7c0042088908338c33d85ce45148565d859561c31d0760810e024969a5c06e5baebaf43a0d75d58b434e7df5d61761c4127f04138e38e39073524a6085465a69e78528e288246ed925989ca29aaab9ee8e4c72c9640b3f7cf1cd37d084134f18824830d03c2451451a41c55863954118a1841a9a19aaa8a342cbadb7e6821df6d8bff3ef7f00358022ca28a4d4720b2e5f9d559b6dff1df824949762aaedb61b836cf2c96297cdf9e7dd7bff3df869b801c734de88434e3a67b1f5565cf8fd17a080401e8964a1b8f6faebbe07376cf4d16acb5df7ddfa33f1841d77e091871ef2dc23134d8111a6dd76dc75e7dd775332dae8a3d2522bf0c01c8f6cf2d5832f5e3cf21b80204218c520a38c33208524d2488d4da6208934e6c8638f924e4a69a50e474cb1c5a7abeeffff37ecd0830f4a2c824c32d63004555452157618620532e8e083a28e4a6aa9e6de8c33cf76c3debfffff3f0145145a5c938d36e38014d257b535e820841c8e48229697629aa9b41d9b8cf6daa9c32e3bfc29bcd0061c7e002208300721d4105072cd45976ec709486082441e2928b2c9dadbf0c416732cf5d6a08feebd073cfc708413a24ce3cf400c3574d4596d79361a79ebb187639f82fa0aacc51d834c72d8628f4db6f4d947308118b6e8c20b31c734f38c53524df518740d4a48a197659a89e6aebcf6eaabc92cbb0c73e189370e39031250e044228a2cc28b3af3d813534e9351569965d76517618d3ed2d9e7aebcf63a2fd0955fceb9fb2ebc00431c78e8410b3e36e1f4d3506cb525da69efc9875f7f649ec9a69cae2abc30c30d3b4c34d4baefee3bf2ffd350830d66a0d1c635dd7813d14c3aed44d767a499965a82104e58619371ca39a7aceaaecb2ebe23935cf2d591239ffcf40928e083119d78020a29da4834d1459559069b6cc3c547df7d295e9925978f1e9b2cb3da768b70c3441f3d77deb0c72efbecf05fd0c10751d491c71eb3d0928b2ef6804452495f81559965cc35e75c7436e2a8e38fa59a7a2aaae49aeb71d663971d79f1c61f8f7c0e40042144127fbc028b2de7a0e31149592176d966a8ad169e785a72d9a5a6aec22a6bbe009f9cf2da749f8e3af5dd831ffef8462cd1c4138820938c33d8189450453ebd05d767a6a156dd78e769b8e1914bca39e7aab0c67a2fd65d7bfdf5e7d0472f7d04545461c5155864a1852bb484238e4a2b9155965967a105997aecb577a194545a59e8afc10e4becc004c75c33da69abbdb6fbef6bf0011d75d871c72bd86cc3cd444111659454800d56d870e599775e84135268619452b6f9e6a8a526fb6cb8e49e1bf1d1a3931e3b1fb9fc724c66b1dda61bb1d36afb2dbffd2edd34e9a53f0f3df8e18baf000e39e800c72fc10843cc3f0625a4905b73b5e69a77e11d9820914532eaa8a79f8e3aedc00a33fcb0e498672e3dfdf89b70c21a6fc8218a38e69cc3d24d6ab1051769a54d671d84114a98e1945a421ae9ace3969baec6218fac74d5565f1d39f0ca337fff0727482186196abcd14932cbb4030f48249dc45460820d469b78e59d47a188243a0965a0820e1aebb8e7a67b31d24a1b8ef8e5986beebe124b30d14440023d15955968b5051781064e79e5b0c41a7bace9a8abbe7a020ae090c3269ce0934f618631069979e775f8e19a6dbac92abb30c72c33da69abed76fcf2cf8f7f2cd258b38d38e86ce4d1565c2dc6d8820f4638619350062a68afc02adc70c71e831cf2e4945b5e7cfaeabb9f4114535851483aeaacf38e4927a1f4946494f5f65b81092ac8608d39ea9826a69aee9aedb9104b3cb1d14b33dd34eebbf3eefbff00a4b042168a8c42ca2cb4e4428e451865b45461861d861871ecb5275f8c32968926a6992adbecc10a3b0c31d043bb0db7e69b577fbdf81250600115904422892ff2cc630f506fc155976fc31587dc81431259a4919d7e0a6ab2ff021c70c154577db5d6b8e7ce3bf01a74f04108308421c6189614630c32fdfc13904e43bd05575ce59db75e7c2ec238638e914e6a29a700076c70c26ec33db7dd83b3de7aedeebf3f02095764b1c517f1c8430f3e228d449255659d95d862df8157de7928a6682596976aba2db8e97afc71c8565f8d75d6b05b8f7df70b30d080034de0914720a970d30d392779d5965c7bfd259861eab907df884f5269e8a1925a9a69b0f6eeeb2fc063935db6d967936efaf2f80fc0420b346ca1c828abc8d2cd38e86c941454554136d96fc529c7dc7efcd988e39d801eba28b2cf764cf4d28e3f0ef9efc23710c10724b040c5155c64e20928a288438e39ea70f45148239975165a69e9865c73d05d379f7d125e58638e5706faa8a49c8a6a2ab0d4729b6fbf26bf4c33ce661f8e78e2bd17ff7df9058c60020b5a7c21c61a9c0c53cc33e850749147229d34545a6ecd55976fbf01175c8434e678e597998a3aeaaae1a65bb1c720efcd77dfa19f3e7d011e98a0020c5a8071c823c824a3cc441879b4d5659e89661a78e4f5f75f8c337e29e6a390e6aa2cb5d9724bb0cd3aebbd37eaae53afbd0137e0c0031b6d4042492dda7033ce451865a491547f0d669879ebc1279f863e0a39a49ecc3e1b6db7e08e0b71d3521baef8f2cd479fbf051760e005249358720d5a6a19d698659b792661a38f1a9cb0e8a4a3deba20831052082e431595d4545561951b79e6a5a7de945566c9a59993b60aebadd25e2b2fc2401b7d34d389035ffc0034e0f0831ca4a0d28b30c8fc3390517f01165871cb39275d751d7e1966999d2a0b6db5dd7e1baec54a2fcdf4dfadbb0e3bf8269c80c20c669c810623bbf0d28b392cb5e4524d6499551863bbf5269e7af4d5771f8433d258a39679eee9e8a4d666cb6db8155f3c34d17b9f3e7cf1d14f7f8002524c414515c518734c325149351555904526996fcf41775f7e228e48e2898f421a69b0e2a6bb30c334d76cf3cd384f5e79e6bb936f7efa17e0b0830f6824f288249c8ce2cc33ecb463cf4f69a905d765c721979c72cb31c7df8a2fc298a5a08416fa68b2ca2e9bb0cb2fc3ecf5d98d3f6e39eecf4b9f3d030d38f080116ec0118926d24c434d35d65c13d04a2d3955155e9865a6d970082eb8a3924e4629a8a29f82baaabbf1ce4bf1c61d978cf2d96867eef9f0c9333f7d002084a0421254e4a14725b8e412cc36ecb4530f4638e554d6598315969a6ada6dc75d77de3598a28a336659a79da0925aaaa9a7b6ebf0c335ebdc35d866a3edf6dbbaf32efcf70620b0800317147104127574f20928cb38c4565b7401179c7006e6a8e38e5e421a29aec13e2b6dc106ff0c74dd76a31e3df6d99f9ffefa1d54b185178ac422cb2cdd0484504229597515565951a65967c599b71e7b12de88638e64fe19e8a1abca4aebaee84e4c31c73c6fddf5d8924f4e79e596abbf3efcf34721c514545461052699f4e28b3aebcc438f4b63b9a51865961d979c7befc137e18c35fa18a4906b5a8ae9a6b38a3b2eb90337fc30c4169b8d76da87b7fe7aecccd78fbffeff0730800143105184118928b2482bbd00034f3c0bb904d34c5761a5d5579d79f699760a9268a28a535279e5968cc65a2baedffa1bf0c07eff0d78e0820f4e38f3f9ebbf7f14b8e4a2cb2e0f4de415599f5d879d77061eb8639067b69929a8a2969aadb6fd0a0cf2c9462bdd36de93535e79f0d66bbf3df9e8a7a0c20a4934e20825c5dc938f3f2431e554555779859863b439f75c7ffefd07608f427a19a69f8a2e0ae9aec07e0beec92f4f4d35df83936ebaf1c7af0f3ffdfaefefff1f831882c838e7a4c30e471fedc453576311561867a209579c72d0c1471f820e3ee8e292514e79e596e08e5b2eba4827adf4d24c371d77dd9253be3bf0e49fffbefcf3a3f0041555d0a1c9269c74b20c33e5b093d0461c8d44545161f97558638eb9461b73cd55179f8311a658e38e3e9689a6a08472dbedc61ea3ac72cb2ed79db7de7cc35efbedb94b7041061b80908516831c128b2ccbf8a31760a7a1366aacb31ecbaebbef523c33cd35db8cb7e597ebce7df9e6b78f7ffefb2bb104134d308249269a00d3cc33d100d491471f4105975c967d865a6aafe517a0801f0aa9e49275027a28adb55eebadbfff9eedb6ffff0310800003e8f2cb461ed9c5975fd1f557a09352f65aecb1ca2a6cf5d558674d3cf2c93b4004124a0c120c31c538538d350adde4d34f663d4659659d89c65c73f555e82188452e99679f808ecaaab2ceb6eb2ec518730c74d045c3edf9e7a0870ec2084d38e10720861cd2082eb97c534e426089455a69b289875e7a28a6a8e2935906faa9a9a7ce7aabaeea128c70c33277edf5d78797ae3af3d363cffdf710b410030d5bfc218821a61c930c3aec5874914d390545d4514b4d7699669c8997de7aedd5d7228c32ceb9eab3d24e1bf0c006272cf6d86497dd38ecb1d7fefd0004207041061b7c5088218728025040020dd494639149461e871e86382595550e9a28afbe029bacc3000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000004feffffffdfffffffffff7fffffffffffffefffffffffffffdfffffffffffff7fffffffffffffffeffffffffffffbffffffffffffbfffffffffffffffeffffffffffffffbffffffffffbfffffffffffffefffffffffffffffdfffffffff7fffffffffeffffffffffdffffffeffffffffffffefffffffffffbffffffffffffefffffffffdffffffffffffffeffffffffff7ffffffffffffffffffefffffffffffffbfffffffffff7fffffffffbffffffffffffdfffffffffffffffbffffffffffff7ffffffffeffffffffffffdfffffffffffbffffffffffbfffffffffffefffffffffff7ffffffffffffdffffffffff7ffffffffffffffff7fffffffffff7ffffffffffbfffffffffffefffffffffffffdfffffffffffffefffffffffffffdfffffffffffefffffffffffffdffffffffffff7ffffffffffffeffffffffffffffeffffffffff7fffffffffffffdffffffffffffffffeffffffffffffbffffffffffffff7fffffffffffffffffffffeffffffffefffffffffff7fffffffffffffbfffffffffffffefffffffffff7fffffffffffffeffffffffffffffbffffffffbffffffffffffdfffffffffffbffffffffffbfffffffffff7ffffffffffffff7fffffffffffdffffffff7ffffffffffffffbfffffffffffbffffffffffeffffffffffefffffffdffffbfffffffff7fffffffffffdffffffffffffbfffffffff7ffffffdffffffffffffffbffffff7fffffffffffeffffffffffdffffffeffffffffff7fffffffffffffeffffffffdffffffffff7ffffffffbffffffffffffbfffffffff7fffffffffffdffffffff7ffffffffffffeffffffff7fffffffffffffdfffffffffffefffffffffffdfffffffffbfffffffffffeffffffffffff7ffffffffffffeffffffffffffffffdfffffffffffffdfffffff7ffffff7ffffffffffffffdffffffffffffbfffffff7fffffffffff7ffffffffff7ffffffffffdfffffffffffffefffffffffffffdffffffffffffffdfffffffffffbffffffffffdfffffffffefffffffffffdfffffffffffffefffffffffff7fffffffffffffeffffffffffffffdfffffffffffbfffffffff7ffffffefffffffffbfffffffffbfffffffff7fffffffff7fffffffffeffffffffffbfffffffffeffffffdffffffffffffdffffffdfffffffffffeffffffffffdfffffffbfffffffff7fffffffffeffffffbfffffffffdfffffffffeffffffffffdfffffffffbffffffbfffffffffffeffffffffffefffffffffdfffffff7fffffffffefffffffffdfffffffffdffffff7fffffffff7ffffffffff7fffffffffdffffffbfffffffffeffffffffffdfffffffffefffffffffdfffffffffeffffffdfffffffffdfffffffffdfffffffffdffffffffffeffffffdffffffffffbfffffffffdffffffefffffffffbffffffffffbfffffffffdfffffff7ffffffffeffffffffffdffffff7fffffffffdffffffffff7fffffffffdfffffffffdfffffffffdfffffffffdfffffffffbfffffffff7fffffffffeffffffeffffffffffbffffff7ffffffffffbffffffefffffffbffffffffffbfffffff7ffffffdfffffffdffffffffffdffffffefffffff7ffffffffffdffffff7fffffffff7ffffffffffffdffffffffefffffffffbfffffffffbffffffffffffdffffffffefffffefffffffffefffffffff7fffffffffffefffffdfffffffbffffffffffbffffffffdffffffffffffbfffffffff7fffffffffffbffffffffdffffffffffff7ffffffffeffffffffffffffffefffffffffffeffffffffdffffffffffffbfffff7fffffffffefffffffffbffffffffffdfffffffffffeffffffffffffffeffffffffffdfffffffffffff7fffffffffffeffffffffffffffbfffffffffffff7fffffffffffffbffffffffdfffffffffffffeffffffffffffff7ffffffffffffefffffffffdffffffffffffbffff7ffffffffffffffdffffffffffeffffffffffffffeffffff033254000000000000000000000b0100000f020000140300001a0400001f050000240600002a0700002f08000034090000390a00003d0b0000420c0000470d00004c0e0000530f0000591000005f110000651200006b13000070140000761500007b16000080170000861800008c190000931a00009a1b0000a11c0000a81d0000ae1e0000b51f0000bc200000c3210000ca220000d2230000d8240000df250000e6260000eb270000f2280000f7290000fc2a0000022c00000000000000000000")
	ef, _ := ReadEliasFano(h)
	_, ok1 := slices.BinarySearch(toArray(ef), 22325642)
	_, ok2 := ef.Search(22325642)
	if ok1 != ok2 {
		panic("why?")
	}
}

func toArray(ef *EliasFano) []uint64 {
	var res []uint64
	it := ef.Iterator()
	for it.HasNext() {
		res = append(res, it.Next())
	}
	return res
}
