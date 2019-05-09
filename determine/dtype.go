package determine

import ( "time")

type( 
  DrillDataType struct{
    OperationList []OperationOne
    SteamCh chan OperationOne
    ScapeDataCh chan ScapeDataD
    ErrCh chan error
    DoneCh chan struct{}
    ScapeFullData bool
    OneMoreScape ScapeDataD
  }

  OperationtypeD [20] string;
  /*
   0 - Бурение
   1 - Наращивание
   2 - Промывка
   3 - Проработка
   4 - Подъем
   5 - Спуск
   6 - Работа т/с
   7 - Бурение ротор
   8 - Бурение (слайд) 
   9 - ПЗР
  */
  ScapeDataD [20] float32;
  /*0-Дата Время
   1-Время Дата
   2=Глубина забоя
   3=Положение долота
   4=Давление на манифольде
   5=Расход на входе
   6=Нагрузка на долото
   7=Вес на крюке
   8=Крутящий момент на роторе
   9=Число оборотов ротора в мин.
   10=Число свечей
   11=Положение долота по свечам
  */ 
  OperationOne struct{
  start,stop time.Time
  startData,stopData,maxData,minData,sum,agv ScapeDataD
  //buf_count,count int;
  //buf:array [0..bufSize] of ageooscape_data;
  Operaton string
  }
)