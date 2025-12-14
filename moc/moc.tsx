import React, { useState } from "react";

const TUIEnglishQuest = () => {
  const [currentScreen, setCurrentScreen] = useState("top");
  const [playerData, setPlayerData] = useState({
    name: "Takuya",
    class: "Vocabulary Warrior",
    level: 2,
    exp: 18,
    maxExp: 50,
    hp: 70,
    maxHp: 100,
    gold: 25,
    streak: 3,
    attack: 16,
    defense: 2.0,
    sessions: 23,
    bestCombo: 8,
  });

  // 単語バトル用のステート
  const [battleState, setBattleState] = useState({
    currentQuestion: 3,
    totalQuestions: 5,
    combo: 2,
    enemyHp: 18,
    enemyMaxHp: 35,
    previousResult: 'Q2: "reduce" → 「減らす」で正解！  +4 経験値, コンボ: 2',
  });

  // 画面切り替え関数
  const navigateTo = (screen) => setCurrentScreen(screen);

  // ヘッダーコンポーネント
  const Header = ({ title, showStatus = false }) => (
    <div className="border-b border-green-500 pb-2 mb-4">
      <div className="text-green-400 font-bold">
        {showStatus ? (
          <>
            TUI English Quest | レベル: {playerData.level} 経験値:{" "}
            {playerData.exp}/{playerData.maxExp}
            HP: {renderHpBar(playerData.hp, playerData.maxHp)} ゴールド:{" "}
            {playerData.gold} 連続日数: {playerData.streak}日
          </>
        ) : (
          `TUI English Quest  |  ${title}`
        )}
      </div>
    </div>
  );

  // フッターコンポーネント
  const Footer = ({ controls }) => (
    <div className="border-t border-green-500 pt-2 mt-4">
      <div className="text-green-400">{controls}</div>
    </div>
  );

  // HPバー描画
  const renderHpBar = (current, max) => {
    const percentage = (current / max) * 10;
    const filled = Math.floor(percentage);
    const empty = 10 - filled;
    return "█".repeat(filled) + "░".repeat(empty);
  };

  // メニュー項目コンポーネント
  const MenuItem = ({ icon, label, description, selected, onClick }) => (
    <div
      className={`cursor-pointer py-1 px-2 ${selected ? "bg-green-900 bg-opacity-30" : ""}`}
      onClick={onClick}
    >
      <span className={selected ? "text-yellow-400" : "text-green-400"}>
        {selected ? "> " : "  "}
        {icon} {label}
      </span>
      {description && <span className="text-gray-400"> - {description}</span>}
    </div>
  );

  // ① トップ画面
  const TopScreen = () => (
    <div className="h-full flex flex-col">
      <Header title="Terminal English RPG" />
      <div className="flex-1 flex flex-col items-center justify-center space-y-8">
        <div className="text-4xl font-bold text-green-400 text-center">
          TUI ENGLISH QUEST
        </div>
        <div className="text-green-300 text-center">
          Learn English by going on an adventure.
        </div>
        <div className="space-y-2 text-center">
          <div
            className="text-green-400 cursor-pointer hover:text-yellow-400"
            onClick={() => navigateTo("home")}
          >
            [ Enter ] 冒険を始める
          </div>
          <div
            className="text-green-400 cursor-pointer hover:text-yellow-400"
            onClick={() => navigateTo("newgame")}
          >
            [ N ] 新規ゲーム
          </div>
          <div className="text-green-400 cursor-pointer hover:text-yellow-400">
            [ Q ] 終了
          </div>
        </div>
      </div>
      <Footer controls="[Enter] 開始  [N] 新規ゲーム  [Q] 終了" />
    </div>
  );

  // ② New Game画面
  const NewGameScreen = () => {
    const [selectedClass, setSelectedClass] = useState(0);
    const classes = [
      {
        name: "Vocabulary Warrior",
        desc: "単語を重視。単語バトルで攻撃力と経験値に小ボーナス。",
      },
      {
        name: "Grammar Mage",
        desc: "文法を重視。文法ダンジョンでダメージ軽減と経験値ボーナス。",
      },
      {
        name: "Conversation Bard",
        desc: "会話を重視。会話タバーンで経験値とゴールドに大ボーナス。",
      },
    ];

    return (
      <div className="h-full flex flex-col">
        <Header title="新規ゲーム" />
        <div className="flex-1 space-y-6">
          <div>
            <div className="text-green-400 mb-2">名前</div>
            <div className="text-yellow-400 ml-4">&gt; Takuya</div>
          </div>

          <div>
            <div className="text-green-400 mb-2">クラス</div>
            <div className="ml-4 space-y-1">
              {classes.map((cls, idx) => (
                <div
                  key={idx}
                  className={`cursor-pointer ${selectedClass === idx ? "text-yellow-400" : "text-green-400"}`}
                  onClick={() => setSelectedClass(idx)}
                >
                  {selectedClass === idx ? "> " : "  "}
                  {cls.name}
                </div>
              ))}
            </div>
          </div>

          <div>
            <div className="text-green-400 mb-2">説明</div>
            <div className="text-gray-300 ml-4">
              {classes[selectedClass].desc}
            </div>
          </div>
        </div>
        <Footer controls="[Tab] フィールド切替  [j/k] 移動  [Enter] 確定  [Esc] 戻る" />
      </div>
    );
  };

  // ③ ホーム画面
  const HomeScreen = () => {
    const [selected, setSelected] = useState(0);
    const menuItems = [
      {
        icon: "⚔",
        label: "単語バトル",
        desc: "単語を5問解いて敵を倒すモード",
        screen: "vocab-battle",
      },
      {
        icon: "🏰",
        label: "文法ダンジョン",
        desc: "文法問題5問でフロア攻略",
        screen: "grammar-dungeon",
      },
      {
        icon: "🍺",
        label: "会話タバーン",
        desc: "NPCと英会話（5ターン）",
        screen: "conversation",
      },
      {
        icon: "🪄",
        label: "スペリングチャレンジ",
        desc: "日本語から英単語をタイプ",
        screen: "spelling",
      },
      {
        icon: "🔊",
        label: "リスニング洞窟",
        desc: "音声を聞いて正しい選択肢を選ぶ",
        screen: "listening",
      },
      {
        icon: "🎒",
        label: "装備",
        desc: "学習ボーナスが付く装備を変更",
        screen: "equipment",
      },
      {
        icon: "🧠",
        label: "弱点AI分析",
        desc: "自分の弱点とおすすめ学習を見る",
        screen: "ai-analysis",
      },
      {
        icon: "📖",
        label: "学習履歴",
        desc: "過去のプレイ結果を確認",
        screen: "history",
      },
      {
        icon: "👤",
        label: "ステータス",
        desc: "レベルやバッジを確認",
        screen: "status",
      },
    ];

    return (
      <div className="h-full flex flex-col">
        <Header showStatus />
        <div className="flex-1">
          <div className="text-green-400 mb-4">どこに行きますか？</div>
          <div className="space-y-0 mb-6">
            {menuItems.map((item, idx) => (
              <MenuItem
                key={idx}
                icon={item.icon}
                label={item.label}
                description={item.desc}
                selected={selected === idx}
                onClick={() => {
                  setSelected(idx);
                  navigateTo(item.screen);
                }}
              />
            ))}
          </div>

          <div className="border border-green-700 p-3 bg-green-950 bg-opacity-30">
            <div className="text-yellow-400 mb-1">ヒント / AIアドバイス</div>
            <div className="text-gray-300 text-sm">
              弱点: 過去形, スペリング
            </div>
            <div className="text-gray-300 text-sm">
              おすすめ: 今日は「スペリングチャレンジ」を2回やってみましょう。
            </div>
          </div>
        </div>
        <Footer controls="[j/k] 移動  [Enter] 決定  [q] 終了" />
      </div>
    );
  };

  // ④ 単語バトル画面
  const VocabBattleScreen = () => {
    const [selected, setSelected] = useState(0);
    const options = ["A. 維持する", "B. 減らす", "C. 投げる", "D. 借りる"];

    return (
      <div className="h-full flex flex-col">
        <div className="border-b border-red-500 pb-2 mb-4">
          <div className="text-red-400 font-bold">
            TUI English Quest | 単語バトル レベル: {playerData.level}
            HP: {renderHpBar(playerData.hp, playerData.maxHp)} コンボ:{" "}
            {battleState.combo}
            敵: スライム ({battleState.enemyHp}/{battleState.enemyMaxHp})
          </div>
        </div>

        <div className="flex-1 space-y-6">
          <div className="text-green-400">
            問題 {battleState.currentQuestion} / {battleState.totalQuestions}
          </div>

          <div>
            <div className="text-green-400 mb-2">英単語</div>
            <div className="text-yellow-400 text-2xl ml-4">"maintain"</div>
          </div>

          <div>
            <div className="text-green-400 mb-3">
              正しい意味を選んでください:
            </div>
            <div className="ml-4 space-y-1">
              {options.map((option, idx) => (
                <div
                  key={idx}
                  className={`cursor-pointer ${selected === idx ? "text-yellow-400" : "text-green-400"}`}
                  onClick={() => setSelected(idx)}
                >
                  {selected === idx ? "> " : "  "}
                  {option}
                </div>
              ))}
            </div>
          </div>
        </div>

        <div className="border-t border-green-700 pt-3 mb-4">
          <div className="text-gray-400 text-sm">前の結果:</div>
          <div className="text-green-300 text-sm ml-2">
            {battleState.previousResult}
          </div>
        </div>

        <Footer controls="[A/B/C/D] 回答  [h] 解説表示（回答後）  [q] 中断" />
      </div>
    );
  };

  // ④-2 バトル結果画面
  const BattleResultScreen = () => (
    <div className="h-full flex flex-col">
      <Header title="単語バトル 結果" />
      <div className="flex-1 space-y-6">
        <div className="text-green-400 text-xl">単語バトル - 結果</div>

        <div className="space-y-2 ml-4">
          <div className="text-green-300">
            正解数 : <span className="text-yellow-400">3 / 5</span>
          </div>
          <div className="text-green-300">
            獲得経験値 : <span className="text-yellow-400">+12</span>
          </div>
          <div className="text-green-300">
            失ったHP : <span className="text-red-400">-20</span>
          </div>
          <div className="text-green-300">
            最高コンボ : <span className="text-yellow-400">2</span>
          </div>
        </div>

        <div className="border border-yellow-700 p-3 bg-yellow-950 bg-opacity-20">
          <div className="text-yellow-400 mb-2">メモ</div>
          <div className="text-gray-300 text-sm mb-1">苦手だった単語:</div>
          <div className="text-gray-300 text-sm ml-4">
            - maintain（維持する）
          </div>
          <div className="text-gray-300 text-sm mt-2">類義語の例:</div>
          <div className="text-gray-300 text-sm ml-4">
            keep, continue, preserve など
          </div>
        </div>
      </div>
      <Footer controls="[Enter] 街に戻る" />
    </div>
  );

  // ⑤ 文法ダンジョン画面
  const GrammarDungeonScreen = () => {
    const [selected, setSelected] = useState(2);
    const options = [
      "A. He don't like coffee.",
      "B. He doesn't likes coffee.",
      "C. He doesn't like coffee.",
      "D. He not like coffee.",
    ];

    return (
      <div className="h-full flex flex-col">
        <div className="border-b border-purple-500 pb-2 mb-4">
          <div className="text-purple-400 font-bold">
            TUI English Quest | 文法ダンジョン フロア: 2/5 レベル:{" "}
            {playerData.level}
            HP: {renderHpBar(playerData.hp, playerData.maxHp)} 防御: 1.2
          </div>
        </div>

        <div className="flex-1 space-y-6">
          <div className="text-red-400 text-lg">文法の罠があらわれた…</div>

          <div>
            <div className="text-green-400 mb-3">
              正しい英文を選んでください:
            </div>
            <div className="ml-4 space-y-1">
              {options.map((option, idx) => (
                <div
                  key={idx}
                  className={`cursor-pointer ${selected === idx ? "text-yellow-400" : "text-green-400"}`}
                  onClick={() => setSelected(idx)}
                >
                  {selected === idx ? "> " : "  "}
                  {option}
                </div>
              ))}
            </div>
          </div>

          <div className="border border-blue-700 p-3 bg-blue-950 bg-opacity-20">
            <div className="text-blue-400 mb-1">解説（回答後に表示）:</div>
            <div className="text-gray-300 text-sm">
              三人称単数 + 現在形 なので「doesn't + 動詞の原形」になります。
              <br />
              He doesn't like coffee. が正しい文です。
            </div>
          </div>

          <div>
            <div className="text-green-400 mb-1">フロア履歴:</div>
            <div className="text-gray-300 text-sm ml-4">
              フロア1: 正解 (+3 経験値, 防御 +0.2)
            </div>
          </div>
        </div>

        <Footer controls="[j/k] 移動  [Enter] 回答  [q] ダンジョンから出る" />
      </div>
    );
  };

  // ⑥ 会話タバーン画面
  const ConversationScreen = () => {
    const [userInput, setUserInput] = useState(
      "I'm heading to the capital city tomorrow.",
    );

    return (
      <div className="h-full flex flex-col">
        <div className="border-b border-orange-500 pb-2 mb-4">
          <div className="text-orange-400 font-bold">
            TUI English Quest | 会話タバーン ゴールド: {playerData.gold}{" "}
            連続日数: {playerData.streak}日
          </div>
        </div>

        <div className="flex-1 space-y-6">
          <div className="border border-orange-700 p-3 bg-orange-950 bg-opacity-20">
            <div className="text-orange-400 mb-2">Old Jaro:</div>
            <div className="text-gray-300 italic ml-4">
              "Hey traveler, what brings you here tonight?"
            </div>
          </div>

          <div>
            <div className="text-green-400 mb-2">あなたの英語での返答:</div>
            <div className="ml-4">
              <div className="text-yellow-400">&gt; {userInput}</div>
            </div>
          </div>

          <div className="border border-green-700 p-3 bg-green-950 bg-opacity-20">
            <div className="text-green-400 mb-2">NPCの返事（送信後）:</div>
            <div className="text-gray-300 italic ml-4">
              "Ah, the capital… busy place. Take the north road and you'll be
              fine."
            </div>
          </div>

          <div>
            <div className="text-green-400 mb-2">セッションの進行状況:</div>
            <div className="text-gray-300 text-sm ml-4 space-y-1">
              <div>ターン: 3 / 5</div>
              <div>現在までの獲得経験値: +10</div>
              <div>獲得ゴールド : +20</div>
            </div>
          </div>
        </div>

        <Footer controls="[Enter] 送信  [↑] 以前の入力を呼び出す  [Esc] 終了" />
      </div>
    );
  };

  // ⑦ スペリングチャレンジ画面
  const SpellingScreen = () => {
    const [userAnswer, setUserAnswer] = useState("maintane");

    return (
      <div className="h-full flex flex-col">
        <div className="border-b border-cyan-500 pb-2 mb-4">
          <div className="text-cyan-400 font-bold">
            TUI English Quest | スペリングチャレンジ 問題: 2/5 HP:{" "}
            {renderHpBar(50, 100)}
          </div>
        </div>

        <div className="flex-1 space-y-6">
          <div>
            <div className="text-green-400 mb-3">
              次の日本語にあてはまる英単語を入力してください:
            </div>
            <div className="text-yellow-400 text-2xl ml-4">「維持する」</div>
          </div>

          <div>
            <div className="text-green-400 mb-2">あなたの回答:</div>
            <div className="text-yellow-400 ml-4">&gt; {userAnswer}</div>
          </div>

          <div className="border border-red-700 p-3 bg-red-950 bg-opacity-20">
            <div className="text-red-400 mb-2">結果:</div>
            <div className="text-gray-300 text-sm mb-2">
              ほぼ正解ですが、スペルが少し違います。
            </div>
            <div className="text-green-300">
              正しいスペル:{" "}
              <span className="text-yellow-400 font-bold">maintain</span>
            </div>
            <div className="text-gray-400 text-sm mt-2">+2 経験値, HP -5</div>
          </div>

          <div className="border border-blue-700 p-3 bg-blue-950 bg-opacity-20">
            <div className="text-blue-400 mb-1">ヒント:</div>
            <div className="text-gray-300 text-sm">
              main + tain の形で覚えるとよいです。
            </div>
          </div>
        </div>

        <Footer controls="[Enter] 次の問題へ  [Esc] 中断" />
      </div>
    );
  };

  // ⑧ リスニング洞窟画面
  const ListeningScreen = () => {
    const [selected, setSelected] = useState(1);
    const options = ["A. Shoes", "B. Coffee", "C. A book", "D. Food"];

    return (
      <div className="h-full flex flex-col">
        <div className="border-b border-indigo-500 pb-2 mb-4">
          <div className="text-indigo-400 font-bold">
            TUI English Quest | リスニング洞窟 問題: 4/5 HP:{" "}
            {renderHpBar(70, 100)}
          </div>
        </div>

        <div className="flex-1 space-y-6">
          <div className="border border-indigo-700 p-4 bg-indigo-950 bg-opacity-30">
            <div className="text-indigo-400 mb-2">音声:</div>
            <div className="text-yellow-400 text-lg ml-4">
              🔊 再生中... "What does she want to buy?"
            </div>
          </div>

          <div>
            <div className="text-green-400 mb-3">
              正しい答えを選んでください:
            </div>
            <div className="ml-4 space-y-1">
              {options.map((option, idx) => (
                <div
                  key={idx}
                  className={`cursor-pointer ${selected === idx ? "text-yellow-400" : "text-green-400"}`}
                  onClick={() => setSelected(idx)}
                >
                  {selected === idx ? "> " : "  "}
                  {option}
                </div>
              ))}
            </div>
          </div>

          <div className="border border-green-700 p-3 bg-green-950 bg-opacity-20">
            <div className="text-green-400 mb-2">結果（回答後）:</div>
            <div className="text-gray-300 text-sm mb-1">正解！</div>
            <div className="text-gray-300 text-sm italic">
              She says: "I'm going to buy some coffee."
            </div>
            <div className="text-yellow-400 text-sm mt-2">+4 経験値</div>
          </div>
        </div>

        <Footer controls="[A/B/C/D] 回答  [r] 音声を再生  [Esc] 中断" />
      </div>
    );
  };

  // ⑨ 装備画面
  const EquipmentScreen = () => {
    const [selected, setSelected] = useState(0);
    const items = [
      { name: "Sword of Words", effect: "+20% 経験値 (単語バトル)" },
      { name: "Shield of Grammar", effect: "-30% ダメージ (文法ダンジョン)" },
      { name: "Ring of Talk", effect: "+50% 経験値 (会話タバーン)" },
      {
        name: "Charm of Spelling",
        effect: "+30% 経験値 (スペリングチャレンジ)",
      },
    ];

    return (
      <div className="h-full flex flex-col">
        <div className="border-b border-amber-500 pb-2 mb-4">
          <div className="text-amber-400 font-bold">
            TUI English Quest | 装備 ゴールド: 120
          </div>
        </div>

        <div className="flex-1 space-y-6">
          <div>
            <div className="text-green-400 mb-3">装備中</div>
            <div className="ml-4 space-y-1 text-sm">
              <div className="text-gray-300">
                武器 : <span className="text-yellow-400">Sword of Words</span>{" "}
                <span className="text-gray-500">
                  (+20% 経験値 in 単語バトル)
                </span>
              </div>
              <div className="text-gray-300">
                防具 :{" "}
                <span className="text-yellow-400">Shield of Grammar</span>{" "}
                <span className="text-gray-500">
                  (-30% ダメージ from 文法ダンジョン)
                </span>
              </div>
              <div className="text-gray-300">
                指輪 : <span className="text-yellow-400">Ring of Talk</span>{" "}
                <span className="text-gray-500">
                  (+50% 経験値 in 会話タバーン)
                </span>
              </div>
              <div className="text-gray-300">
                お守り:{" "}
                <span className="text-yellow-400">Charm of Spelling</span>{" "}
                <span className="text-gray-500">
                  (+30% 経験値 in スペリング)
                </span>
              </div>
            </div>
          </div>

          <div>
            <div className="text-green-400 mb-3">インベントリ</div>
            <div className="ml-4 space-y-1">
              {items.map((item, idx) => (
                <div
                  key={idx}
                  className={`cursor-pointer ${selected === idx ? "text-yellow-400" : "text-green-400"}`}
                  onClick={() => setSelected(idx)}
                >
                  {selected === idx ? "> " : "  "}
                  {item.name}
                  <span className="text-gray-500 text-sm ml-2">
                    {item.effect}
                  </span>
                </div>
              ))}
            </div>
          </div>
        </div>

        <Footer controls="[j/k] 移動  [Enter] 装備/解除  [Esc] 街に戻る" />
      </div>
    );
  };

  // ⑩ AI分析画面
  const AIAnalysisScreen = () => (
    <div className="h-full flex flex-col">
      <Header title="弱点AI分析" />
      <div className="flex-1 space-y-6">
        <div className="text-green-400 text-lg">
          あなたの最近のパフォーマンス（過去80問）
        </div>

        <div className="border border-red-700 p-3 bg-red-950 bg-opacity-20">
          <div className="text-red-400 mb-2">弱点:</div>
          <div className="text-gray-300 text-sm ml-4 space-y-1">
            <div>- 過去形</div>
            <div>- 前置詞</div>
            <div>- スペリング（-tain / -tion で終わる単語）</div>
          </div>
        </div>

        <div className="border border-green-700 p-3 bg-green-950 bg-opacity-20">
          <div className="text-green-400 mb-2">得意分野:</div>
          <div className="text-gray-300 text-sm ml-4 space-y-1">
            <div>- 基本的な日常単語</div>
            <div>- 現在形の文法</div>
          </div>
        </div>

        <div className="border border-yellow-700 p-3 bg-yellow-950 bg-opacity-20">
          <div className="text-yellow-400 mb-2">おすすめ:</div>
          <div className="text-gray-300 text-sm ml-4 space-y-1">
            <div>- スペリングチャレンジを今日2回プレイ</div>
            <div>- 過去形に焦点を当てた文法ダンジョンを1回</div>
            <div>- 会話タバーンで過去形の文を使ってみる</div>
          </div>
        </div>

        <div>
          <div className="text-green-400 mb-2">今日の推奨プラン:</div>
          <div className="text-gray-300 text-sm ml-4 space-y-1">
            <div>1. スペリングチャレンジ (1セッション)</div>
            <div>2. 文法ダンジョン (1セッション)</div>
            <div>3. 会話タバーン (1セッション)</div>
          </div>
        </div>
      </div>
      <Footer controls="[Enter] OK  [Esc] 街に戻る" />
    </div>
  );

  // ⑪ ステータス画面
  const StatusScreen = () => (
    <div className="h-full flex flex-col">
      <Header title="ステータス" />
      <div className="flex-1 space-y-6">
        <div>
          <div className="text-green-400 mb-2">プレイヤー</div>
          <div className="text-gray-300 text-sm ml-4 space-y-1">
            <div>
              名前 : <span className="text-yellow-400">{playerData.name}</span>
            </div>
            <div>
              クラス:{" "}
              <span className="text-yellow-400">{playerData.class}</span>
            </div>
          </div>
        </div>

        <div>
          <div className="text-green-400 mb-2">ステータス</div>
          <div className="text-gray-300 text-sm ml-4 space-y-1">
            <div>
              レベル :{" "}
              <span className="text-yellow-400">{playerData.level}</span>
            </div>
            <div>
              経験値 :{" "}
              <span className="text-yellow-400">
                {playerData.exp} / {playerData.maxExp}
              </span>
            </div>
            <div>
              HP :{" "}
              <span className="text-yellow-400">
                {playerData.hp} / {playerData.maxHp}
              </span>
            </div>
            <div>
              攻撃力 :{" "}
              <span className="text-yellow-400">{playerData.attack}</span>
            </div>
            <div>
              防御力 :{" "}
              <span className="text-yellow-400">{playerData.defense}</span>
            </div>
          </div>
        </div>

        <div>
          <div className="text-green-400 mb-2">進行状況</div>
          <div className="text-gray-300 text-sm ml-4 space-y-1">
            <div>
              連続日数 :{" "}
              <span className="text-yellow-400">{playerData.streak} 日</span>
            </div>
            <div>
              最高コンボ :{" "}
              <span className="text-yellow-400">{playerData.bestCombo}</span>
            </div>
            <div>
              プレイセッション:{" "}
              <span className="text-yellow-400">{playerData.sessions}</span>
            </div>
          </div>
        </div>

        <div>
          <div className="text-green-400 mb-2">バッジ</div>
          <div className="text-gray-300 text-sm ml-4 space-y-1">
            <div>✓ Sharp Mind (10問連続正解)</div>
            <div className="text-gray-600">
              🔒 Consistency Medal (7日連続ログイン)
            </div>
            <div className="text-yellow-600">
              ⏳ Vocabulary Knight (50単語マスター) [進行中]
            </div>
          </div>
        </div>
      </div>
      <Footer controls="[Esc] 街に戻る" />
    </div>
  );

  // ⑫ 履歴画面
  const HistoryScreen = () => {
    const [selected, setSelected] = useState(0);
    const sessions = [
      {
        icon: "⚔",
        name: "単語バトル",
        date: "2025-12-10",
        result: "3/5 正解",
        exp: "+12",
        hp: "-20",
      },
      {
        icon: "🏰",
        name: "文法ダンジョン",
        date: "2025-12-10",
        result: "4/5 正解",
        exp: "+15",
        hp: "-6",
      },
      {
        icon: "🍺",
        name: "会話タバーン",
        date: "2025-12-09",
        result: "5/5 ターン",
        exp: "+18",
        gold: "+40",
      },
      {
        icon: "🪄",
        name: "スペリングチャレンジ",
        date: "2025-12-09",
        result: "2/5 完璧",
        exp: "+10",
        hp: "-15",
      },
      {
        icon: "🔊",
        name: "リスニング洞窟",
        date: "2025-12-08",
        result: "4/5 正解",
        exp: "+14",
        hp: "-6",
      },
    ];

    return (
      <div className="h-full flex flex-col">
        <Header title="学習履歴" />
        <div className="flex-1 space-y-6">
          <div className="text-green-400">最近のセッション</div>

          <div className="space-y-2">
            {sessions.map((session, idx) => (
              <div
                key={idx}
                className={`cursor-pointer p-2 ${selected === idx ? "bg-green-900 bg-opacity-30" : ""}`}
                onClick={() => setSelected(idx)}
              >
                <div
                  className={
                    selected === idx ? "text-yellow-400" : "text-green-400"
                  }
                >
                  {selected === idx ? "> " : "  "}
                  {session.icon} {session.name}
                  <span className="text-gray-500 ml-4">{session.date}</span>
                </div>
                <div className="text-gray-400 text-sm ml-6">
                  {session.result} {session.exp} {session.hp || session.gold}
                </div>
              </div>
            ))}
          </div>

          <div className="text-gray-400 text-sm">
            セッションを選択すると詳細を確認できます
          </div>
        </div>
        <Footer controls="[j/k] 移動  [Enter] 詳細  [Esc] 街に戻る" />
      </div>
    );
  };

  // ⑬ 戦闘不能画面
  const FaintedScreen = () => (
    <div className="h-full flex flex-col">
      <Header title="戦闘不能..." />
      <div className="flex-1 flex flex-col items-center justify-center space-y-6">
        <div className="text-red-400 text-3xl">💀</div>
        <div className="text-red-400 text-xl">
          ダメージを受けすぎて倒れてしまった
        </div>

        <div className="text-gray-300 text-center space-y-2">
          <div className="text-red-300">経験値 -5</div>
          <div className="text-green-300">HPが50%回復しました</div>
        </div>

        <div className="text-yellow-400 text-center">
          続けましょう！小さな失敗は旅の一部です。
        </div>
      </div>
      <Footer controls="[Enter] 街に戻る" />
    </div>
  );

  // ⑭ レベルアップ画面
  const LevelUpScreen = () => (
    <div className="h-full flex flex-col">
      <Header title="レベルアップ！" />
      <div className="flex-1 flex flex-col items-center justify-center space-y-6">
        <div className="text-yellow-400 text-4xl">⭐</div>
        <div className="text-yellow-400 text-2xl">レベル 3 に到達！</div>

        <div className="text-gray-300 space-y-2">
          <div className="text-green-300">
            最大HP <span className="text-yellow-400">+10</span>
          </div>
          <div className="text-green-300">
            攻撃力 <span className="text-yellow-400">+2</span>
          </div>
          <div className="text-green-300">
            防御力 <span className="text-yellow-400">+1</span>
          </div>
        </div>

        <div className="border border-cyan-700 p-3 bg-cyan-950 bg-opacity-20">
          <div className="text-cyan-400 mb-1">
            新しいチャレンジが解放されました:
          </div>
          <div className="text-gray-300 text-sm ml-4">
            - リスニング洞窟（初級モード）
          </div>
        </div>
      </div>
      <Footer controls="[Enter] 続ける" />
    </div>
  );

  // 画面ルーティング
  const renderScreen = () => {
    switch (currentScreen) {
      case "top":
        return <TopScreen />;
      case "newgame":
        return <NewGameScreen />;
      case "home":
        return <HomeScreen />;
      case "vocab-battle":
        return <VocabBattleScreen />;
      case "battle-result":
        return <BattleResultScreen />;
      case "grammar-dungeon":
        return <GrammarDungeonScreen />;
      case "conversation":
        return <ConversationScreen />;
      case "spelling":
        return <SpellingScreen />;
      case "listening":
        return <ListeningScreen />;
      case "equipment":
        return <EquipmentScreen />;
      case "ai-analysis":
        return <AIAnalysisScreen />;
      case "status":
        return <StatusScreen />;
      case "history":
        return <HistoryScreen />;
      case "fainted":
        return <FaintedScreen />;
      case "levelup":
        return <LevelUpScreen />;
      default:
        return <HomeScreen />;
    }
  };

  return (
    <div className="min-h-screen bg-black text-green-400 font-mono p-4">
      <div className="max-w-5xl mx-auto h-screen flex flex-col">
        {renderScreen()}
      </div>

      {/* クイックナビゲーション（デモ用） */}
      <div className="fixed bottom-4 right-4 bg-gray-900 border border-gray-700 p-3 rounded text-xs space-y-1">
        <div className="text-gray-500 mb-2">画面切替（デモ用）:</div>
        <button
          onClick={() => navigateTo("top")}
          className="block w-full text-left text-green-400 hover:text-yellow-400"
        >
          トップ
        </button>
        <button
          onClick={() => navigateTo("home")}
          className="block w-full text-left text-green-400 hover:text-yellow-400"
        >
          ホーム
        </button>
        <button
          onClick={() => navigateTo("vocab-battle")}
          className="block w-full text-left text-green-400 hover:text-yellow-400"
        >
          単語バトル
        </button>
        <button
          onClick={() => navigateTo("battle-result")}
          className="block w-full text-left text-green-400 hover:text-yellow-400"
        >
          バトル結果
        </button>
        <button
          onClick={() => navigateTo("grammar-dungeon")}
          className="block w-full text-left text-green-400 hover:text-yellow-400"
        >
          文法ダンジョン
        </button>
        <button
          onClick={() => navigateTo("conversation")}
          className="block w-full text-left text-green-400 hover:text-yellow-400"
        >
          会話タバーン
        </button>
        <button
          onClick={() => navigateTo("spelling")}
          className="block w-full text-left text-green-400 hover:text-yellow-400"
        >
          スペリング
        </button>
        <button
          onClick={() => navigateTo("listening")}
          className="block w-full text-left text-green-400 hover:text-yellow-400"
        >
          リスニング
        </button>
        <button
          onClick={() => navigateTo("equipment")}
          className="block w-full text-left text-green-400 hover:text-yellow-400"
        >
          装備
        </button>
        <button
          onClick={() => navigateTo("ai-analysis")}
          className="block w-full text-left text-green-400 hover:text-yellow-400"
        >
          AI分析
        </button>
        <button
          onClick={() => navigateTo("status")}
          className="block w-full text-left text-green-400 hover:text-yellow-400"
        >
          ステータス
        </button>
        <button
          onClick={() => navigateTo("history")}
          className="block w-full text-left text-green-400 hover:text-yellow-400"
        >
          履歴
        </button>
        <button
          onClick={() => navigateTo("fainted")}
          className="block w-full text-left text-green-400 hover:text-yellow-400"
        >
          戦闘不能
        </button>
        <button
          onClick={() => navigateTo("levelup")}
          className="block w-full text-left text-green-400 hover:text-yellow-400"
        >
          レベルアップ
        </button>
      </div>
    </div>
  );
};

export default TUIEnglishQuest;
